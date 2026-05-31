// Package withdraw 출금 기능 슬라이스 (TS의 application/withdraw + in/kafka/handlers/withdraw 대응)
// Kafka 출금 명령 수신 → 번들러/블록체인으로 전송 → 결과 발행
//
// TODO(골격): account 슬라이스 패턴(account.go/service.go/store.go/handler.go) 따라 구현
package withdraw

// Kafka 토픽 — 출금 (방향은 BC Adapter 기준: In=수신, Out=발행)
const (
	TopicRequest   = "adapter.withdraw.request"   // In  출금 요청
	TopicResult    = "adapter.withdraw.result"    // Out 출금 결과
	TopicStatus    = "adapter.withdraw.status"    // In  출금 상태 확인
	TopicConfirmed = "adapter.withdraw.confirmed" // Out 출금 최종 결과
)

// Asset 전송 자산 (payment.Asset 과 동일 구조 — 슬라이스 독립을 위해 별도 정의)
type Asset struct {
	Type    string  `json:"type"`    // NATIVE / ERC20
	Address *string `json:"address"` // ERC20 컨트랙트 주소, NATIVE는 null
	Amount  string  `json:"amount"`  // 최소 단위 정수 문자열
	Symbol  *string `json:"symbol"`
}

// ResultError 처리 실패 사유 (성공 시 code/message 모두 null)
type ResultError struct {
	Code    *string `json:"code"`
	Message *string `json:"message"`
}

// Request 출금 요청 페이로드(adapter.withdraw.request 인바운드)
//
// Message Key = from_address(파티션 내 nonce 순서 보장)
type Request struct {
	WTradeNo    string `json:"w_trade_no"`
	ChainID     int64  `json:"chain_id"`
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	Asset       Asset  `json:"asset"`
}

// Result 출금 결과 페이로드(adapter.withdraw.result 아웃바운드)
type Result struct {
	WTradeNo         string       `json:"w_trade_no"`
	Status           string       `json:"status"` // SUBMITTED / CONFIRMED / FAILED
	ChainID          int64        `json:"chain_id"`
	UserOpHash       string       `json:"userop_hash"`
	TxHash           *string      `json:"tx_hash"`
	CompleteDatetime string       `json:"complete_datetime"` // yyyyMMddHHmmss
	Error            *ResultError `json:"error,omitempty"`
}

// StatusRequest 출금 상태 확인 요청 페이로드(adapter.withdraw.status 인바운드)
//
// TODO(설계): 설계서에 페이로드 샘플 없음 — confirm 요청에서 추론, 구현 전 확정
type StatusRequest struct {
	WTradeNo string `json:"w_trade_no"`
	TxHash   string `json:"tx_hash"`
	ChainID  int64  `json:"chain_id"`
}

// Confirmed 출금 최종 결과 페이로드(adapter.withdraw.confirmed 아웃바운드)
//
// TODO(설계): 설계서에 페이로드 샘플 없음 — confirm 결과에서 추론, 구현 전 확정
type Confirmed struct {
	WTradeNo     string `json:"w_trade_no"`
	TxHash       string `json:"tx_hash"`
	Status       string `json:"status"` // CONFIRMED / FAILED
	ChainID      int64  `json:"chain_id"`
	ConfirmCount int    `json:"confirm_count"`
}
