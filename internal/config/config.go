// Package config 체인/토큰 설정 등록 기능 슬라이스
// (TS의 application/config-register + in/kafka/handlers/config-register 대응)
// 주의: 애플리케이션 env 설정은 platform/env, 이건 "체인/토큰 설정 등록" 도메인
//
// TODO(설계): 토픽·페이로드 정의 출처 없음 — 컨벤션 추론 초안, 등록 필드 미정, WalletBE 계약 확정 필요
// TODO(골격): account 슬라이스 패턴 따라 구현
package config

// Kafka 토픽 — 체인/토큰 설정 등록 (방향은 BC Adapter 기준: In=수신, Out=발행)
const (
	TopicRegister   = "adapter.config.register"   // In  설정 등록 요청
	TopicRegistered = "adapter.config.registered" // Out 설정 등록 결과
)

// RegisterCommand 체인/토큰 설정 등록 명령(adapter.config.register 인바운드)
//
// TODO(설계): RPC URL·토큰 컨트랙트·심볼·decimals 등 등록 필드 미정
type RegisterCommand struct {
	RequestID string `json:"request_id"`
	ChainID   int64  `json:"chain_id"`
}

// Registered 설정 등록 결과(adapter.config.registered 아웃바운드)
type Registered struct {
	RequestID string `json:"request_id"`
	ChainID   int64  `json:"chain_id"`
	Success   bool   `json:"success"`
}
