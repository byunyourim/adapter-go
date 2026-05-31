// Package settlementcollect 정산 수금 기능 슬라이스
// (TS의 application/settlement-collect + in/kafka/handlers/settlement-collect 대응)
//
// TODO(설계): 아래 토픽·페이로드는 설계서(부록 B)·TS 원본 어디에도 정의가 없어
// adapter.<도메인>.<동작> 컨벤션과 유사 토픽을 참고한 초안이다. 수금 대상·금액 등
// 핵심 필드가 미정이므로 WalletBE 실제 계약 확정 전까지 신뢰하지 말 것
// (토픽명 settlement.collect 표기도 확정 대상 — settlement_collect 가능성)
//
// TODO(골격): account 슬라이스 패턴(service.go/store.go/handler.go) 따라 구현
package settlementcollect

// Kafka 토픽 — 정산 수금 (방향은 BC Adapter 기준: In=수신, Out=발행)
const (
	TopicRequest = "adapter.settlement.collect.request" // In  정산 수금 요청
	TopicResult  = "adapter.settlement.collect.result"  // Out 정산 수금 결과
)

// Request 정산 수금 요청 페이로드(adapter.settlement.collect.request 인바운드)
//
// TODO(설계): 수금 대상·금액·자산 등 필드 미정
type Request struct {
	RequestID   string `json:"request_id"`
	ChainID     int64  `json:"chain_id"`
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
}

// Result 정산 수금 결과 페이로드(adapter.settlement.collect.result 아웃바운드)
type Result struct {
	RequestID string  `json:"request_id"`
	Status    string  `json:"status"` // SUBMITTED / CONFIRMED / FAILED
	ChainID   int64   `json:"chain_id"`
	TxHash    *string `json:"tx_hash"`
}
