module github.com/byunyourim/stablecoinbc-adapter

go 1.25

// 리스너(StableCoinBC_Adapter_Listener)와 버전 통일.
// 공유 의존성은 반드시 동일 버전을 사용한다 — 상위 stablecoin/go.work + `go work sync`로 정렬.
//   go-ethereum   v1.17.3
//   gorilla/websocket v1.4.2 (go-ethereum 간접)
// 아직 양쪽에 미추가된 의존성(추가 시 두 모듈 동일 버전 유지):
//   jackc/pgx, segmentio/kafka-go, redis/go-redis, caarlos0/env
// 어댑터 전용: entgo.io/ent (ORM, pgx 드라이버 백엔드) — 리스너는 sqlc라 미사용
require github.com/ethereum/go-ethereum v1.17.3
