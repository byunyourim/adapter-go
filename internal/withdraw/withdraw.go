// Package withdraw 출금 기능 슬라이스 (TS의 application/withdraw + in/kafka/handlers/withdraw 대응)
// Kafka 출금 명령 수신 → 번들러/블록체인으로 전송 → 결과 발행
//
// TODO(골격): account 슬라이스 패턴(account.go/service.go/store.go/handler.go) 따라 구현
package withdraw
