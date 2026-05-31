// Package confirm 트랜잭션 확정(컨펌) 기능 슬라이스
// (TS의 application/confirm + in/kafka/handlers/confirm 대응)
//
// deposit.confirm / common.confirm 공용 페이로드를 이 패키지에서 정의
//
// TODO(골격): account 슬라이스 패턴 따라 구현
package confirm

// Kafka 토픽 — 컨펌 확인 (방향은 BC Adapter 기준: In=수신, Out=발행)
const (
	TopicDepositConfirm   = "adapter.deposit.confirm"   // In  입금 컨펌 요청
	TopicDepositConfirmed = "adapter.deposit.confirmed" // Out 입금 컨펌 결과
	TopicCommonConfirm    = "adapter.common.confirm"    // In  공통 컨펌 요청
	TopicCommonConfirmed  = "adapter.common.confirmed"  // Out 공통 컨펌 결과
)

// Request 컨펌 확인 요청 페이로드(deposit.confirm / common.confirm 공용 인바운드)
type Request struct {
	WTradeNo string `json:"w_trade_no"`
	TxHash   string `json:"tx_hash"`
	ChainID  int64  `json:"chain_id"`
}

// Result 컨펌 확인 결과 페이로드(deposit.confirmed / common.confirmed 공용 아웃바운드)
type Result struct {
	WTradeNo     string `json:"w_trade_no"`
	TxHash       string `json:"tx_hash"`
	Status       string `json:"status"` // CONFIRMED / PENDING_CONFIRMATION / FAILED
	ChainID      int64  `json:"chain_id"`
	ConfirmCount int    `json:"confirm_count"`
}
