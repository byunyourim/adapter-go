// Package confirm 트랜잭션 확정(컨펌) 기능 슬라이스
// (TS의 application/confirm + in/kafka/handlers/confirm 대응)
//
// deposit.confirm 과 common.confirm 은 동일한 요청/결과 페이로드를 쓰므로
// 한 곳(이 패키지)에서 정의한다
//
// TODO(골격): account 슬라이스 패턴(service.go/store.go/handler.go) 따라 구현
package confirm

// Kafka 토픽 — 컨펌 확인 (방향은 BC Adapter 기준: In=수신, Out=발행)
const (
	TopicDepositConfirm   = "adapter.deposit.confirm"   // In  입금 컨펌 요청
	TopicDepositConfirmed = "adapter.deposit.confirmed" // Out 입금 컨펌 결과
	TopicCommonConfirm    = "adapter.common.confirm"    // In  공통 컨펌 요청
	TopicCommonConfirmed  = "adapter.common.confirmed"  // Out 공통 컨펌 결과
)

// Request 컨펌 확인 요청 페이로드(deposit.confirm / common.confirm 공용 인바운드)
//
// 설계서 3.5.3 기준. confirmations = latestBlockNumber - txBlockNumber + 1
type Request struct {
	WTradeNo string `json:"w_trade_no"`
	TxHash   string `json:"tx_hash"`
	ChainID  int64  `json:"chain_id"`
}

// Result 컨펌 확인 결과 페이로드(deposit.confirmed / common.confirmed 공용 아웃바운드)
//
// 기준 컨펌 수 충족 여부 판단은 WalletBE 책임(체인/자산별 min_confirmations)
type Result struct {
	WTradeNo     string `json:"w_trade_no"`
	TxHash       string `json:"tx_hash"`
	Status       string `json:"status"` // CONFIRMED / PENDING_CONFIRMATION / FAILED
	ChainID      int64  `json:"chain_id"`
	ConfirmCount int    `json:"confirm_count"`
}
