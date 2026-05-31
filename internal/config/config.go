// Package config 체인/토큰 설정 등록 기능 슬라이스
// (TS의 application/config-register + in/kafka/handlers/config-register 대응)
// 주의: 애플리케이션 env 설정은 platform/env, 이건 "체인/토큰 설정 등록" 도메인
//
// TODO(설계): 아래 토픽·페이로드는 설계서(부록 B)·TS 원본 어디에도 정의가 없어
// account.create/created 의 동사쌍 컨벤션을 참고한 초안이다. 등록할 체인 RPC·토큰
// 컨트랙트 등 필드가 미정이므로 WalletBE 실제 계약 확정 전까지 신뢰하지 말 것
//
// TODO(골격): account 슬라이스 패턴(service.go/store.go/handler.go) 따라 구현
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
