// Package kafka Kafka producer/consumer 제공
// (TS의 adapter/out/kafka + adapter/in/kafka 대응) segmentio/kafka-go 사용
package kafka

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/segmentio/kafka-go"
)

// Producer 결과/이벤트를 Kafka로 발행
type Producer struct {
	w *kafka.Writer
}

// NewProducer 주어진 브로커로 발행 Producer 생성
//
// Murmur2 파티셔너로 kafkajs(TS) 기본 파티셔닝과 호환 — 같은 key는 같은 파티션에
// 들어가 순서가 보장된다. acks=all 로 유실을 막는다(at-least-once)
func NewProducer(brokers []string) *Producer {
	return &Producer{
		w: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Balancer:     &kafka.Murmur2Balancer{},
			RequiredAcks: kafka.RequireAll,
		},
	}
}

// Publish 토픽에 메시지 1건 발행. key는 파티션 라우팅(순서 보장)에 사용
func (p *Producer) Publish(ctx context.Context, topic string, key, payload []byte) error {
	err := p.w.WriteMessages(ctx, kafka.Message{Topic: topic, Key: key, Value: payload})
	if err != nil {
		return fmt.Errorf("kafka publish(%s): %w", topic, err)
	}
	return nil
}

// Close 내부 writer 종료
func (p *Producer) Close() error {
	return p.w.Close()
}

// HandlerFunc 한 메시지를 처리하는 함수 — 기능 슬라이스의 핸들러로 등록
type HandlerFunc func(ctx context.Context, payload []byte) error

// Publisher DLQ 발행에 필요한 최소 인터페이스 — *Producer 가 구현, 테스트는 대역 주입
type Publisher interface {
	Publish(ctx context.Context, topic string, key, payload []byte) error
}

// Consumer 토픽별 핸들러로 메시지를 디스패치
//
// at-least-once: 핸들러 성공 후에만 offset commit. 핸들러가 에러를 반환하면 해당
// 메시지를 DLQ로 격리한 뒤 commit 한다(재시도해도 동일 실패하는 메시지가 파티션을
// 막는 것을 방지). 일시적 실패의 재시도는 핸들러 내부 책임 — 에러 반환 시점엔
// 재처리 불가로 간주한다
type Consumer struct {
	brokers  []string
	groupID  string
	handlers map[string]HandlerFunc
	dlq      Publisher // nil이면 DLQ 비활성(실패 메시지는 commit 막아 재처리)
	log      *slog.Logger
}

// NewConsumer consumer 생성. dlq는 실패 메시지 발행 대상(nil 허용)
func NewConsumer(brokers []string, groupID string, dlq Publisher, log *slog.Logger) *Consumer {
	return &Consumer{
		brokers:  brokers,
		groupID:  groupID,
		handlers: make(map[string]HandlerFunc),
		dlq:      dlq,
		log:      log,
	}
}

// Register 토픽에 핸들러 등록 (TS의 handlers/register 대응)
func (c *Consumer) Register(topic string, h HandlerFunc) {
	c.handlers[topic] = h
}

// Run 등록된 토픽들을 구독해 메시지를 수신하고 핸들러로 디스패치
//
// ctx 취소 시 정상 종료(nil). 등록 핸들러가 없으면 즉시 에러
func (c *Consumer) Run(ctx context.Context) error {
	if len(c.handlers) == 0 {
		return errors.New("kafka consumer: 등록된 핸들러 없음")
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     c.brokers,
		GroupID:     c.groupID,
		GroupTopics: c.topics(),
		StartOffset: kafka.LastOffset, // fromBeginning=false 대응(신규 그룹은 최신부터)
	})
	defer r.Close()

	for {
		m, err := r.FetchMessage(ctx)
		if err != nil {
			if isShutdown(err) {
				return nil
			}
			return fmt.Errorf("kafka fetch: %w", err)
		}

		if err := c.handle(ctx, m); err != nil {
			if isShutdown(err) {
				return nil
			}
			// DLQ 미설정/발행 실패 — commit 없이 종료해 재처리(at-least-once 보전)
			return err
		}

		if err := r.CommitMessages(ctx, m); err != nil {
			if isShutdown(err) {
				return nil
			}
			return fmt.Errorf("kafka commit(%s): %w", m.Topic, err)
		}
	}
}

// handle 메시지 1건을 핸들러로 디스패치. commit해도 되면 nil, 재처리/종료가 필요하면 에러
func (c *Consumer) handle(ctx context.Context, m kafka.Message) error {
	h, ok := c.handlers[m.Topic]
	if !ok {
		// 구독 토픽엔 핸들러가 있어야 정상 — 방어적으로 DLQ
		return c.toDLQ(ctx, m, fmt.Errorf("핸들러 미등록: %s", m.Topic))
	}
	if err := h(ctx, m.Value); err != nil {
		if isShutdown(err) {
			return err
		}
		c.log.Error("메시지 처리 실패 → DLQ", "topic", m.Topic, "err", err)
		return c.toDLQ(ctx, m, err)
	}
	return nil
}

// toDLQ 실패 메시지를 "<원토픽>.DLQ"로 발행. DLQ 미설정/발행 실패면 에러(commit 막음)
func (c *Consumer) toDLQ(ctx context.Context, m kafka.Message, cause error) error {
	if c.dlq == nil {
		return fmt.Errorf("처리 실패(DLQ 미설정): %w", cause)
	}
	if err := c.dlq.Publish(ctx, m.Topic+".DLQ", m.Key, m.Value); err != nil {
		return fmt.Errorf("DLQ 발행 실패(원인 %v): %w", cause, err)
	}
	return nil
}

// topics 등록된 토픽 목록
func (c *Consumer) topics() []string {
	ts := make([]string, 0, len(c.handlers))
	for t := range c.handlers {
		ts = append(ts, t)
	}
	return ts
}

// isShutdown ctx 취소/마감을 셧다운 신호로 판별 — 재시도/에러 전파 안 함
func isShutdown(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}
