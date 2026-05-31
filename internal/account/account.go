// Package account 입금 지갑 계좌의 생명주기(생성/배포/삭제) 기능 슬라이스
//
// 이 패키지가 "지갑 배포(CREATE2)" 담당. 배포는 온체인 트랜잭션 확정을
// 기다리는 느린 작업이라, 요청 경로에서 분리해 Kafka 명령 → 비동기 처리
// 동일 디플로이어 계정 nonce 직렬화는 Redis 분산락 + nonce 재시도로 격리
// (README "지갑 배포 비동기화" 참고)
package account

import (
	"errors"
	"math/big"
)

// Kafka 토픽 — 계정 생명주기 (방향은 BC Adapter 기준: In=수신, Out=발행)
const (
	TopicCreate   = "adapter.account.create"   // In  계정 생성 요청
	TopicCreated  = "adapter.account.created"  // Out 계정 생성 결과
	TopicDeploy   = "adapter.account.deploy"   // In  계정 배포 요청
	TopicDeployed = "adapter.account.deployed" // Out 계정 배포 결과
)

// Account 입금 지갑 계좌 도메인 모델
type Account struct {
	ChainID      int64
	Address      string // EIP-55 checksum
	Salt         string // CREATE2 salt
	Deployed     bool
	DeployTxHash string
	PredictedAt  string
}

// DeployCommand 지갑 배포 명령(Kafka 인바운드)
type DeployCommand struct {
	ChainID int64  `json:"chain_id"`
	Address string `json:"address"`
	Salt    string `json:"salt"`
}

// validate 명령 필수 필드 검증
func (c DeployCommand) validate() error {
	if c.ChainID == 0 {
		return errors.New("chainId is required")
	}
	if c.Address == "" {
		return errors.New("address is required")
	}
	if c.Salt == "" {
		return errors.New("salt is required")
	}
	return nil
}

// DeployResult 배포 처리 결과(Kafka 아웃바운드)
type DeployResult struct {
	ChainID int64    `json:"chain_id"`
	Address string   `json:"address"`
	TxHash  string   `json:"tx_hash"`
	Gas     *big.Int `json:"gas,omitempty"`
	Success bool     `json:"success"`
}

// AccountCreateRequest 계정 생성 요청 페이로드(adapter.account.create 인바운드)
//
// 설계서 3.1.3 기준. Message Key = userId(사용자 단위 순서 보장)
// 생성 시 온체인 트랜잭션 없이 CREATE2 주소만 사전 계산(배포는 첫 출금/입금 시 lazy)
type AccountCreateRequest struct {
	RequestID   string `json:"request_id"`
	NetworkType string `json:"network_type"`
}

// AccountCreatedResult 계정 생성 결과 페이로드(adapter.account.created 아웃바운드)
type AccountCreatedResult struct {
	RequestID string `json:"request_id"`
	Address   string `json:"address"` // EIP-55 checksum 사전 계산 주소
}
