package account

import (
	"context"
	"encoding/json"
	"fmt"
)

// KafkaHandler account 관련 Kafka 메시지를 디코드해 Service로 라우팅
// (TS의 adapter/in/kafka/handlers/account-*.handler.ts 대응)
//
// 인바운드 어댑터는 역직렬화/검증만 담당하고 도메인 로직은 Service에 위임.
type KafkaHandler struct {
	svc *Service
}

// NewKafkaHandler 핸들러 생성
func NewKafkaHandler(svc *Service) *KafkaHandler {
	return &KafkaHandler{svc: svc}
}

// HandleDeployMessage raw 메시지를 DeployCommand로 파싱해 처리
// platform/kafka.HandlerFunc 시그니처(func(ctx, []byte) error)와 호환
func (h *KafkaHandler) HandleDeployMessage(ctx context.Context, payload []byte) error {
	var cmd DeployCommand
	if err := json.Unmarshal(payload, &cmd); err != nil {
		// 파싱 불가 메시지는 재시도해도 동일 실패 — consumer가 DLQ로 보내야 한다.
		return fmt.Errorf("배포 메시지 파싱 실패: %w", err)
	}
	if err := cmd.validate(); err != nil {
		return fmt.Errorf("배포 메시지 검증 실패: %w", err)
	}
	return h.svc.HandleDeploy(ctx, cmd)
}
