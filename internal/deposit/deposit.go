// Package deposit 는 Listener로부터 WebSocket으로 입금 이벤트를 수신하는
// 인바운드 슬라이스다. (TS의 adapter/in/websocket/deposit-event-handler 대응)
//
// 첫 입금 감지 시 지갑 미배포 계좌면 account 슬라이스에 배포 명령을 던지고(비동기),
// 입금 자체 처리는 막지 않는다.
package deposit

import "context"

// Event 는 Listener가 보낸 입금 이벤트.
type Event struct {
	ChainID   int64
	TxHash    string
	LogIndex  int
	ToAddress string
	Amount    string
	Symbol    string
	Status    string // TXCF / TXPD
}

// DeployRequester 는 미배포 지갑의 배포를 비동기로 요청한다(account 슬라이스로 위임).
type DeployRequester interface {
	RequestDeploy(ctx context.Context, chainID int64, address string) error
}

// Handler 는 수신한 입금 이벤트를 처리한다.
type Handler struct {
	deploy DeployRequester
}

// NewHandler 는 핸들러를 만든다.
func NewHandler(deploy DeployRequester) *Handler {
	return &Handler{deploy: deploy}
}

// Handle 는 입금 이벤트 1건을 처리한다.
//
// TODO(골격): 멱등 확인 → 입금 적재 → 미배포면 deploy.RequestDeploy(비동기)
func (h *Handler) Handle(ctx context.Context, e Event) error {
	panic("not implemented")
}
