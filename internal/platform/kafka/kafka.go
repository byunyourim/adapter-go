// Package kafka Kafka producer/consumer 제공
// (TS의 adapter/out/kafka + adapter/in/kafka 대응) segmentio/kafka-go 사용
package kafka

import "context"

// Producer 결과/이벤트 발행
type Producer struct {
	// TODO(골격): *kafka.Writer
}

// NewProducer producer 생성
//
// TODO(골격): broker 주입
func NewProducer(broker string) *Producer {
	panic("not implemented")
}

// Publish 토픽에 메시지 발행
func (p *Producer) Publish(ctx context.Context, topic string, key, payload []byte) error {
	panic("not implemented")
}

// HandlerFunc 한 메시지를 처리하는 함수 — 기능 슬라이스의 핸들러로 등록
type HandlerFunc func(ctx context.Context, payload []byte) error

// Consumer 토픽별 핸들러로 메시지 디스패치
// at-least-once: 핸들러 성공 후에만 offset commit. 실패 시 재시도/DLQ.
type Consumer struct {
	handlers map[string]HandlerFunc
}

// NewConsumer consumer 생성
func NewConsumer() *Consumer {
	return &Consumer{handlers: make(map[string]HandlerFunc)}
}

// Register 토픽에 핸들러 등록 (TS의 handlers/register 대응)
func (c *Consumer) Register(topic string, h HandlerFunc) {
	c.handlers[topic] = h
}

// Run 메시지를 수신해 등록된 핸들러로 디스패치
//
// TODO(골격): kafka-go Reader 루프 + offset commit + DLQ
func (c *Consumer) Run(ctx context.Context) error {
	panic("not implemented")
}
