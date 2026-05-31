// Package settlement 정산 기능 슬라이스 (TS의 application/settlement + in/kafka/handlers/settlement 대응)
//
// TODO(설계): 토픽·페이로드 정의 출처 없음 — 컨벤션 추론 초안, 핵심 필드 미정, WalletBE 계약 확정 필요
// TODO(골격): account 슬라이스 패턴 따라 구현
package settlement

// Kafka 토픽 — 정산 (In=수신, Out=발행)
const (
	TopicRequest = "adapter.settlement.request" // In  정산 요청
	TopicResult  = "adapter.settlement.result"  // Out 정산 결과
)

// Request 정산 요청 페이로드
//
// TODO(설계): 정산 대상(가맹점/기간)·금액·자산 등 필드 미정
type Request struct {
	RequestID string `json:"request_id"`
	ChainID   int64  `json:"chain_id"`
}

// Result 정산 결과 페이로드
type Result struct {
	RequestID string  `json:"request_id"`
	Status    string  `json:"status"` // SUBMITTED / CONFIRMED / FAILED
	ChainID   int64   `json:"chain_id"`
	TxHash    *string `json:"tx_hash"`
}
