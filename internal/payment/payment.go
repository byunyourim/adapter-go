// Package payment 결제 기능 슬라이스 (TS의 application/payment + in/kafka/handlers/payment 대응)
//
// TODO(골격): account 슬라이스 패턴 따라 구현
package payment

// Kafka 토픽 — 결제 (방향은 BC Adapter 기준: In=수신, Out=발행)
const (
	TopicRequest = "adapter.payment.request" // In  결제 요청
	TopicResult  = "adapter.payment.result"  // Out 결제 결과
)

// Asset 결제 자산 (withdraw.Asset 과 동일 구조 — 슬라이스 독립을 위해 별도 정의)
type Asset struct {
	Type    string  `json:"type"`    // NATIVE / ERC20
	Address *string `json:"address"` // ERC20 컨트랙트 주소, NATIVE는 null
	Amount  string  `json:"amount"`
	Symbol  *string `json:"symbol"`
}

// ResultError 처리 실패 사유 (성공 시 code/message 모두 null)
type ResultError struct {
	Code    *string `json:"code"`
	Message *string `json:"message"`
}

// Request 결제 요청 페이로드(adapter.payment.request 인바운드)
type Request struct {
	RequestID   string  `json:"request_id"`
	TraceID     string  `json:"trace_id"`
	ChainID     int64   `json:"chain_id"`
	FromAddress *string `json:"from_address"`
	ToAddress   *string `json:"to_address"`
	Asset       Asset   `json:"asset"`
}

// Result 결제 결과 페이로드(adapter.payment.result 아웃바운드)
//
// TODO(설계): 설계서에 결과 페이로드 샘플 없음 — withdraw.Result 에서 추론, 구현 전 확정
type Result struct {
	RequestID        string       `json:"request_id"`
	Status           string       `json:"status"` // REQUESTED / PROCESSING / SUBMITTED / PENDING / CONFIRMED / FAILED
	ChainID          int64        `json:"chain_id"`
	UserOpHash       string       `json:"userop_hash"`
	TxHash           *string      `json:"tx_hash"`
	CompleteDatetime string       `json:"complete_datetime"` // yyyyMMddHHmmss
	Error            *ResultError `json:"error,omitempty"`
}
