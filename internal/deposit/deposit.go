// Package deposit Listener로부터 WebSocket으로 입금 이벤트를 수신하는
// 인바운드 슬라이스 (TS의 adapter/in/websocket/deposit-event-handler 대응)
//
// 첫 입금 감지 시 지갑 미배포 계좌면 account 슬라이스에 배포 명령을 던지고(비동기),
// 입금 자체 처리는 막지 않음
package deposit

import "context"

// TopicDetected 입금 감지 알림 토픽(아웃바운드, BC Adapter → WalletBE)
const TopicDetected = "adapter.deposit.detected"

// Event Listener가 보낸 입금 이벤트(WebSocket 인바운드, 내부 표현)
type Event struct {
	ChainID   int64
	TxHash    string
	LogIndex  int
	ToAddress string
	Amount    string
	Symbol    string
	Status    string // TXCF / TXPD
}

// Detected 입금 감지 알림 페이로드(adapter.deposit.detected 아웃바운드)
//
// 멱등 키 = (tx_hash, log_index), 멱등성 보장은 WalletBE 책임
type Detected struct {
	ChainID             int64  `json:"chain_id"`
	TxHash              string `json:"tx_hash"`
	FromAddress         string `json:"from_address"`
	ToAddress           string `json:"to_address"`
	Amount              string `json:"amount"`        // 최소 단위 정수 문자열
	Status              string `json:"status"`        // DETECTED / PENDING_CONFIRMATION / CONFIRMED / FAILED
	ConfirmCount        string `json:"confirm_count"` // 설계서상 문자열("10")
	Symbol              string `json:"symbol"`
	TransactionDatetime string `json:"transaction_datetime"` // yyyyMMddHHmmss
}

// DeployRequester 미배포 지갑 배포를 비동기 요청 (account 슬라이스로 위임)
type DeployRequester interface {
	RequestDeploy(ctx context.Context, chainID int64, address string) error
}

// Handler 수신한 입금 이벤트 처리
type Handler struct {
	deploy DeployRequester
}

// NewHandler 핸들러 생성
func NewHandler(deploy DeployRequester) *Handler {
	return &Handler{deploy: deploy}
}

// Handle 입금 이벤트 1건 처리
//
// TODO(골격): 멱등 확인 → 입금 적재 → 미배포면 deploy.RequestDeploy(비동기)
func (h *Handler) Handle(ctx context.Context, e Event) error {
	panic("not implemented")
}
