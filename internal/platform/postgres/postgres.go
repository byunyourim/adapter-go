// Package postgres 는 Ent 클라이언트(pgx 드라이버 기반)를 제공한다.
// (TS의 adapter/out/persistence/database 대응)
// 각 도메인 패키지의 Store가 ent.Client를 주입받아 쓴다.
// Ent 스키마는 ent/schema/ 에 코드로 정의하고 `go generate`로 코드 생성.
package postgres

import "context"

// Pool 은 pgx 풀 래퍼.
type Pool struct {
	// TODO(골격): *pgxpool.Pool
}

// New 는 DSN으로 풀을 연다.
//
// TODO(골격): pgxpool.New + Ping.
func New(ctx context.Context, dsn string) (*Pool, error) {
	panic("not implemented")
}

// Close 는 풀을 닫는다.
func (p *Pool) Close() {
	panic("not implemented")
}
