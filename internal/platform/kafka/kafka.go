// Package kafka 는 Kafka producer/consumer를 제공한다.
// (TS의 adapter/out/kafka + adapter/in/kafka 대응) segmentio/kafka-go 사용.
package kafka

import "context"

// Producer 는 결과/이벤트를 발행한다.
type Producer struct {
	// TODO(골격): *kafka.Writer
}

// NewProducer 는 producer를 만든다.
//
// TODO(골격): broker 주입.
func NewProducer(broker string) *Producer {
	panic("not implemented")
}

// Publish 는 토픽에 메시지를 발행한다.
func (p *Producer) Publish(ctx context.Context, topic string, key, payload []byte) error {
	panic("not implemented")
}

// HandlerFunc 는 한 메시지를 처리하는 함수. 기능 슬라이스의 핸들러를 등록한다.
type HandlerFunc func(ctx context.Context, payload []byte) error

// Consumer 는 토픽별 핸들러로 메시지를 디스패치한다.
// at-least-once: 핸들러 성공 후에만 offset commit. 실패 시 재시도/DLQ.
type Consumer struct {
	handlers map[string]HandlerFunc
}

// NewConsumer 는 consumer를 만든다.
func NewConsumer() *Consumer {
	return &Consumer{handlers: make(map[string]HandlerFunc)}
}

// Register 는 토픽에 핸들러를 등록한다. (TS의 handlers/register 대응)
func (c *Consumer) Register(topic string, h HandlerFunc) {
	c.handlers[topic] = h
}

// Run 은 메시지를 수신해 등록된 핸들러로 디스패치한다.
//
// TODO(골격): kafka-go Reader 루프 + offset commit + DLQ.
func (c *Consumer) Run(ctx context.Context) error {
	panic("not implemented")
}
