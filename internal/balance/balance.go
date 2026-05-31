// Package balance 잔액 조회 기능 슬라이스
// (TS의 application/balance + in/kafka/handlers/balance 대응)
//
// TODO(설계): 토픽·페이로드 정의 출처 없음 — 컨벤션 추론 초안, WalletBE 계약 확정 필요
// TODO(골격): account 슬라이스 패턴 따라 구현
package balance

// Kafka 토픽 — 잔액 조회 (방향은 BC Adapter 기준: In=수신, Out=발행)
const (
	TopicRequest = "adapter.balance.request" // In  잔액 조회 요청
	TopicResult  = "adapter.balance.result"  // Out 잔액 조회 결과
)

// Request 잔액 조회 요청 페이로드(adapter.balance.request 인바운드)
type Request struct {
	RequestID string `json:"request_id"`
	ChainID   int64  `json:"chain_id"`
	Address   string `json:"address"`
}

// Holding 자산별 보유 잔액
type Holding struct {
	Symbol       string  `json:"symbol"`
	Amount       string  `json:"amount"`        // 최소 단위 정수 문자열
	TokenAddress *string `json:"token_address"` // ERC20 컨트랙트, NATIVE는 null
}

// Result 잔액 조회 결과 페이로드(adapter.balance.result 아웃바운드)
type Result struct {
	RequestID string    `json:"request_id"`
	ChainID   int64     `json:"chain_id"`
	Address   string    `json:"address"`
	Holdings  []Holding `json:"holdings"`
}
