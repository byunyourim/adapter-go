// Command adapter 는 입금/출금/정산/지갑 배포를 오케스트레이션하는 서비스다.
//
// 인바운드: Listener로부터 WebSocket 입금 이벤트, Kafka 명령(출금/정산/계좌 등)
// 아웃바운드: 블록체인(go-ethereum), 외부 번들러, Kafka 결과 발행, Redis 분산락
//
// 아키텍처(패키지 by 기능 + 이벤트 드리븐)와 지갑 배포 비동기화는 README.md 참고.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	log.Info("adapter starting")

	// TODO(골격): wiring 순서
	//  1. env.Load()
	//  2. platform: postgres.Pool / kafka.{Producer,Consumer} / redis.Locker / blockchain.Client / bundler.Client
	//  3. 기능 service 조립 (account, withdraw, payment, settlement, ...)
	//  4. 인바운드 등록:
	//       - deposit.NewWSHandler(...)        ← Listener WS 수신
	//       - kafka consumer에 기능별 handler 라우팅 등록
	//  5. 서버/컨슈머 start
	<-ctx.Done()

	log.Info("adapter shutting down")
	// TODO(골격): graceful shutdown — 컨슈머 정지, in-flight 완료, 연결 종료
}
