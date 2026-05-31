package kafka

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeDLQ Publisher 대역 — DLQ 발행 호출을 기록
type fakeDLQ struct {
	calls []kafka.Message
	err   error
}

func (f *fakeDLQ) Publish(_ context.Context, topic string, key, payload []byte) error {
	f.calls = append(f.calls, kafka.Message{Topic: topic, Key: key, Value: payload})
	return f.err
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// TestConsumerHandle handle의 at-least-once/DLQ 분기를 고정
func TestConsumerHandle(t *testing.T) {
	okHandler := func(context.Context, []byte) error { return nil }
	failHandler := func(context.Context, []byte) error { return errors.New("boom") }
	cancelHandler := func(context.Context, []byte) error { return context.Canceled }

	tests := []struct {
		name        string
		topic       string // 수신 메시지 토픽
		register    string // 핸들러 등록 토픽("" = 등록 안 함)
		handler     HandlerFunc
		withDLQ     bool
		dlqErr      error
		wantErr     bool   // handle이 에러 반환(=commit 막음/종료)
		wantDLQ     bool   // DLQ로 보냈는지
		wantDLQName string // DLQ 토픽명
	}{
		{
			name:  "성공이면 commit(에러 없음), DLQ 미호출",
			topic: "adapter.account.deploy", register: "adapter.account.deploy",
			handler: okHandler, withDLQ: true,
			wantErr: false, wantDLQ: false,
		},
		{
			name:  "핸들러 실패 + DLQ면 DLQ로 격리 후 commit",
			topic: "adapter.account.deploy", register: "adapter.account.deploy",
			handler: failHandler, withDLQ: true,
			wantErr: false, wantDLQ: true, wantDLQName: "adapter.account.deploy.DLQ",
		},
		{
			name:  "핸들러 실패 + DLQ 없으면 에러(commit 막음)",
			topic: "adapter.account.deploy", register: "adapter.account.deploy",
			handler: failHandler, withDLQ: false,
			wantErr: true, wantDLQ: false,
		},
		{
			name:  "미등록 토픽은 방어적으로 DLQ",
			topic: "adapter.unknown", register: "",
			handler: okHandler, withDLQ: true,
			wantErr: false, wantDLQ: true, wantDLQName: "adapter.unknown.DLQ",
		},
		{
			name:  "ctx 취소는 셧다운 — 에러 전파, DLQ 미호출",
			topic: "adapter.account.deploy", register: "adapter.account.deploy",
			handler: cancelHandler, withDLQ: true,
			wantErr: true, wantDLQ: false,
		},
		{
			name:  "DLQ 발행 실패면 에러(commit 막음)",
			topic: "adapter.account.deploy", register: "adapter.account.deploy",
			handler: failHandler, withDLQ: true, dlqErr: errors.New("dlq down"),
			wantErr: true, wantDLQ: true, wantDLQName: "adapter.account.deploy.DLQ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var dlq *fakeDLQ
			var pub Publisher
			if tt.withDLQ {
				dlq = &fakeDLQ{err: tt.dlqErr}
				pub = dlq
			}
			c := NewConsumer([]string{"localhost:9092"}, "test-group", pub, discardLogger())
			if tt.register != "" {
				c.Register(tt.register, tt.handler)
			}

			err := c.handle(context.Background(), kafka.Message{
				Topic: tt.topic, Key: []byte("k"), Value: []byte("v"),
			})

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if !tt.wantDLQ {
				if dlq != nil {
					assert.Empty(t, dlq.calls, "DLQ로 보내면 안 됨")
				}
				return
			}
			require.Len(t, dlq.calls, 1, "DLQ 1건 발행 기대")
			assert.Equal(t, tt.wantDLQName, dlq.calls[0].Topic)
			assert.Equal(t, []byte("v"), dlq.calls[0].Value, "원본 payload 보존")
			assert.Equal(t, []byte("k"), dlq.calls[0].Key, "원본 key 보존")
		})
	}
}

// TestConsumerRunNoHandlers 핸들러 미등록 시 즉시 에러
func TestConsumerRunNoHandlers(t *testing.T) {
	c := NewConsumer([]string{"localhost:9092"}, "test-group", nil, discardLogger())
	err := c.Run(context.Background())
	require.Error(t, err)
}
