// Package redis 는 분산락을 제공한다. (TS의 adapter/out/redis/redis-lock + distributed-lock.port 대응)
// 지갑 배포 시 디플로이어 계정 단위 락으로 nonce 직렬화를 격리하는 핵심 컴포넌트.
package redis

import (
	"context"
	"time"
)

// Locker 는 분산락. account.Locker 인터페이스를 만족한다.
type Locker struct {
	ttl time.Duration
	// TODO(골격): *redis.Client
}

// NewLocker 는 Locker를 만든다.
//
// TODO(골격): go-redis 클라이언트 주입.
func NewLocker(ttl time.Duration) *Locker {
	return &Locker{ttl: ttl}
}

// WithLock 은 key 락을 잡고 fn을 실행한 뒤 해제한다.
//
// TODO(골격): SET NX PX + 토큰 기반 해제(Lua), 재진입 방지.
func (l *Locker) WithLock(ctx context.Context, key string, fn func() error) error {
	panic("not implemented")
}
