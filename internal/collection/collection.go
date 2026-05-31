// Package collection 입금 자금 회수 기능 슬라이스
// (TS의 application/collection 대응)
//
// TODO(설계): 아래 토픽·페이로드는 설계서(부록 B)·TS 원본 어디에도 정의가 없어
// adapter.<도메인>.<동작> 컨벤션과 유사 토픽(withdraw)을 참고한 초안이다. WalletBE
// 실제 계약 확정 전까지 토픽명·필드를 신뢰하지 말 것
//
// TODO(골격): account 슬라이스 패턴(service.go/store.go/handler.go) 따라 구현
package collection

// Kafka 토픽 — 입금 자금 회수 (방향은 BC Adapter 기준: In=수신, Out=발행)
const (
	TopicRequest = "adapter.collection.request" // In  자금 회수 요청
	TopicResult  = "adapter.collection.result"  // Out 자금 회수 결과
)

// Request 입금 자금 회수 요청 페이로드(adapter.collection.request 인바운드)
//
// TODO(설계): 회수 대상·금액·자산 등 필드 미정
type Request struct {
	RequestID   string `json:"request_id"`
	ChainID     int64  `json:"chain_id"`
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
}

// Result 입금 자금 회수 결과 페이로드(adapter.collection.result 아웃바운드)
type Result struct {
	RequestID string  `json:"request_id"`
	Status    string  `json:"status"` // SUBMITTED / CONFIRMED / FAILED
	ChainID   int64   `json:"chain_id"`
	TxHash    *string `json:"tx_hash"`
}
