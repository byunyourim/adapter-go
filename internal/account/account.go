// Package account 는 입금 지갑 계좌의 생명주기(생성/배포/삭제) 기능 슬라이스다.
//
// 이 패키지가 "지갑 배포(CREATE2)"를 담당한다. 배포는 온체인 트랜잭션 확정을
// 기다리는 느린 작업이므로, 요청 경로에서 분리해 Kafka 명령 → 비동기 처리한다.
// 동일 디플로이어 계정의 nonce 직렬화는 Redis 분산락 + nonce 재시도로 격리한다.
// (README "지갑 배포 비동기화" 참고)
package account

import (
	"errors"
	"math/big"
)

// Account 는 입금 지갑 계좌 도메인 모델.
type Account struct {
	ChainID      int64
	Address      string // EIP-55 checksum
	Salt         string // CREATE2 salt
	Deployed     bool
	DeployTxHash string
	PredictedAt  string
}

// DeployCommand 는 지갑 배포 명령(Kafka 인바운드).
type DeployCommand struct {
	ChainID int64  `json:"chainId"`
	Address string `json:"address"`
	Salt    string `json:"salt"`
}

// validate 는 명령 필수 필드를 검증한다.
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

// DeployResult 는 배포 처리 결과(Kafka 아웃바운드).
type DeployResult struct {
	ChainID int64    `json:"chainId"`
	Address string   `json:"address"`
	TxHash  string   `json:"txHash"`
	Gas     *big.Int `json:"gas,omitempty"`
	Success bool     `json:"success"`
}
