// Package settlementcollect 정산 수금 기능 슬라이스
// (TS의 application/settlement-collect + in/kafka/handlers/settlement-collect 대응)
//
// TODO(설계): 토픽·페이로드 정의 출처 없음 — 컨벤션 추론 초안(토픽명 settlement.collect 확정 대상), WalletBE 계약 확정 필요
// TODO(골격): account 슬라이스 패턴 따라 구현
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
