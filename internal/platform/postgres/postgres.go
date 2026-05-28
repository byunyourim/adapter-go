// Package postgres Ent 클라이언트(pgx 드라이버 기반) 제공
// (TS의 adapter/out/persistence/database 대응)
// 각 도메인 패키지의 Store가 ent.Client를 주입받아 사용
// Ent 스키마는 ent/schema/ 에 코드로 정의하고 `go generate`로 코드 생성
package postgres

import "context"

// Pool pgx 풀 래퍼
type Pool struct {
	// TODO(골격): *pgxpool.Pool
}

// New DSN으로 풀 연결
//
// TODO(골격): pgxpool.New + Ping
func New(ctx context.Context, dsn string) (*Pool, error) {
	panic("not implemented")
}

// Close 풀 닫기
func (p *Pool) Close() {
	panic("not implemented")
}
