// Package deposit Listener로부터 WebSocket으로 입금 이벤트를 수신하는
// 인바운드 슬라이스 (TS의 adapter/in/websocket/deposit-event-handler 대응)
//
// 첫 입금 감지 시 지갑 미배포 계좌면 account 슬라이스에 배포 명령을 던지고(비동기),
// 입금 자체 처리는 막지 않음
package deposit

import "context"

// Event Listener가 보낸 입금 이벤트
type Event struct {
	ChainID   int64
	TxHash    string
	LogIndex  int
	ToAddress string
	Amount    string
	Symbol    string
	Status    string // TXCF / TXPD
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
