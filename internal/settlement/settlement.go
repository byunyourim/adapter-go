// Package settlement 정산 기능 슬라이스 (TS의 application/settlement + in/kafka/handlers/settlement 대응)
//
// TODO(설계): 아래 토픽·페이로드는 설계서(부록 B)·TS 원본 어디에도 정의가 없어
// adapter.<도메인>.<동작> 컨벤션과 유사 토픽을 참고한 초안이다. 정산 대상·금액 등
// 핵심 필드가 미정이므로 WalletBE 실제 계약 확정 전까지 신뢰하지 말 것
//
// TODO(골격): account 슬라이스 패턴(service.go/store.go/handler.go) 따라 구현
package settlement

// Kafka 토픽 — 정산 (방향은 BC Adapter 기준: In=수신, Out=발행)
const (
	TopicRequest = "adapter.settlement.request" // In  정산 요청
	TopicResult  = "adapter.settlement.result"  // Out 정산 결과
)

// Request 정산 요청 페이로드(adapter.settlement.request 인바운드)
//
// TODO(설계): 정산 대상(가맹점/기간)·금액·자산 등 필드 미정
type Request struct {
	RequestID string `json:"request_id"`
	ChainID   int64  `json:"chain_id"`
}

// Result 정산 결과 페이로드(adapter.settlement.result 아웃바운드)
type Result struct {
	RequestID string  `json:"request_id"`
	Status    string  `json:"status"` // SUBMITTED / CONFIRMED / FAILED
	ChainID   int64   `json:"chain_id"`
	TxHash    *string `json:"tx_hash"`
}
