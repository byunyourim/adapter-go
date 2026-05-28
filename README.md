# StableCoinBC Adapter (Go)

입금/출금/정산/지갑 배포를 오케스트레이션하는 코어 서비스. Listener로부터 입금 이벤트를
받고, Kafka 명령을 처리하며, 블록체인·외부 번들러와 통신한다. 기존 TypeScript 버전을
Go로 재작성한 프로젝트.

---

## 아키텍처: 패키지 by 기능 + 이벤트 드리븐

헥사고날(port/adapter 4겹)을 그대로 이식하지 않고, **Go에서 가장 보편적인 "도메인별 납작
패키지"** 로 재구성했다. 기술 레이어(domain/application/adapter/infra)로 나누지 않고,
`internal/` 바로 아래에 도메인 패키지를 두며 인프라만 `platform/`으로 묶는다.

```
cmd/adapter/main.go            # 단일 진입점 — 모든 패키지 wiring

internal/
  # ── 도메인 패키지 (각 패키지가 자기 타입+서비스+저장소+핸들러를 보유) ──
  account/                     # 계좌 생성/삭제/배포(CREATE2) ★ 지갑 배포 문제
    account.go                 #   도메인 타입
    service.go                 #   유스케이스 + 소비자측 인터페이스(Locker/Deployer/...)
    deploy.go                  #   배포 처리 (분산락 + nonce 격리)
    store.go                   #   Postgres 저장소
    handler.go                 #   Kafka 인바운드 핸들러
  deposit/                     # Listener WS 입금 수신 (인바운드)
  withdraw/ payment/ settlement/ settlementcollect/
  collection/ confirm/ balance/ config/
  chain/                       # 공유: 체인/토큰 모델, 심볼 정규화
  evm/                         # 공유: EIP-55, CREATE2 주소 계산, ERC-20/UserOp 헬퍼

  # ── 플랫폼(인프라 클라이언트). 도메인 패키지는 좁은 인터페이스로만 의존 ──
  platform/
    postgres/ kafka/ redis/ blockchain/ bundler/ env/ logger/ apperr/

migrations/                    # Postgres 마이그레이션
```

### 왜 이 구조인가

- **Go 보편 컨벤션**: 레이어 분리(domain/app/infra) 없이 도메인 패키지를 납작하게 두는 게 Go에서 가장 흔한 방식. 디렉토리 1개 = 패키지 1개로 분리는 충분하다.
- **응집도**: 출금을 고치려면 `withdraw/` 한 폴더만 본다.
- **Go 관용 인터페이스**: 인터페이스를 **사용하는 쪽에서 좁게** 선언한다(`account/service.go`의 `Locker`/`Deployer`). 구현체(`platform/*`)는 implicit하게 만족. 자바식 "구현체 옆 큰 인터페이스"를 버린 형태.
- **의존성 방향**: `도메인 패키지 → platform 인터페이스`, `→ chain/evm`. 역방향·패키지 간 직접 의존 금지(필요하면 이벤트로). 강제하려면 `depguard`(golangci-lint) 규칙 추가.
- **이벤트 드리븐**: 느린 작업(지갑 배포 등)은 Kafka 명령으로 비동기 처리. → 아래 참고.

> 트레이드오프: 이 방식은 도메인 타입과 저장소(store.go)를 한 패키지에 둔다(도메인-인프라 co-location). 도메인 격리를 더 엄격히 원하면 레이어드(domain/app/platform)로 갈 수 있으나, Go 보편 컨벤션을 우선해 납작 구조를 택했다.

---

## 지갑 배포 비동기화 (이 프로젝트의 핵심 개선)

**문제**: 첫 입금 시 CREATE2 지갑을 온체인 배포하는데, 배포는 트랜잭션 확정까지 느리고,
같은 디플로이어 계정을 쓰는 다른 요청이 nonce 순서 때문에 그 뒤에 줄 서서 전체가 느려진다.

**해결 (구조로)**:

1. **요청 경로에서 분리** — 입금 감지(`deposit/`)는 미배포 계좌면 배포를 *명령*으로 던지고 즉시 다음 처리로 넘어간다. 배포 완료를 기다리지 않는다.
2. **Kafka 비동기 처리** — `account/deploy.go`가 배포 명령을 컨슈머로 받아 처리. 입금/조회 등 다른 요청과 격리된 흐름.
3. **nonce 직렬화 격리** — 같은 디플로이어 계정의 배포끼리만 **Redis 분산락**(`platform/redis`)으로 순서를 보장하고, 다른 계정/체인 배포는 병렬 진행. nonce는 `platform/blockchain`이 원자적으로 관리.
4. **멱등성** — at-least-once(Kafka 재배달) 대비. 이미 배포된 주소면 skip. 입금 주소는 CREATE2로 배포 전에도 예측 가능하므로(`evm.PredictCreate2`) 입금 수신 자체는 배포와 무관하게 가능.

> 핵심: **"배포는 요청 경로에서 떼어내고, 직렬화는 디플로이어 단위 락으로 좁게 가둔다."**

---

## 기술 스택

### 확정 스택

| 영역 | 선택 | 이유 |
|------|------|------|
| 언어 | Go 1.25 | 진짜 멀티스레드, 단일 바이너리, 블록체인 백엔드 1군 |
| 체인 연동 | go-ethereum + abigen | ethers의 Go 표준 대응 |
| DB 엔진 | PostgreSQL | SQLite 단일 쓰기 락 제거, 트랜잭션·동시성 |
| DB 드라이버 | jackc/pgx | Ent 백엔드 드라이버. Postgres 전용 최고 성능, 풀 내장 |
| ORM | **Ent** | 엔티티/관계 풍부한 어댑터 도메인에 적합. 타입 안전 코드 생성, 명시적 트랜잭션 |
| 메시지 큐 | segmentio/kafka-go | 기존 kafkajs 대응. 컨슈머 그룹·offset 제어 |
| 캐시/락 | redis/go-redis | 기존 ioredis 대응. 분산락(배포 nonce 격리) |
| 로깅 | log/slog (표준) | 외부 의존 0, 구조화 |
| 설정 | caarlos0/env | struct 태그 매핑(@ConfigurationProperties 감각) |
| 마이그레이션 | Ent migrate (atlas) | 스키마를 Ent 코드로 정의 → 자동 마이그레이션 |

### 선택지가 있는 스택 (장단점 + 선택 근거)

#### DB 접근 방식 — **Ent** (ORM) + 돈 핵심 경로는 raw SQL

리스너는 단순 read뿐이라 sqlc로 충분하지만, 어댑터는 account·settlement·payment 등
**엔티티/관계가 풍부**해 ORM 생산성이 실이득이다.

| 후보 | 장점 | 단점 |
|------|------|------|
| **Ent** ✅ | 스키마 코드 정의→타입 안전 생성, 관계/그래프 강력, **트랜잭션 명시적**(`client.Tx`), raw SQL escape hatch | 학습·코드젠 |
| sqlc | SQL 그대로·투명 | 관계 많은 CRUD에 보일러플레이트 ↑ (리스너엔 적합) |
| GORM | JPA 감각, CRUD 생산성 | 런타임 리플렉션, 암묵적 트랜잭션 — **돈 도메인에 부적합** |

**선택 이유**: 어댑터의 풍부한 도메인엔 ORM 생산성이 필요하다. 단 GORM의 암묵적 동작은
돈 처리에 위험하므로, **트랜잭션이 명시적이고 코드 생성형(매직 없음)인 Ent**를 택한다.
JPA 생산성은 살리되 "언제 commit되는지"가 코드에 드러난다.

**하이브리드**: 엔티티 CRUD 대부분은 Ent로, 출금/정산처럼 **돈이 걸린 핵심 트랜잭션**은
Ent의 raw SQL(또는 직접 pgx)로 명시 제어한다. 한 프로젝트에서 혼용 가능.

#### 메시지 처리 — **at-least-once + 멱등성**

at-most-once(유실 위험)가 아니라 **at-least-once**(핸들러 성공 후 offset commit) + 멱등 처리. 돈 처리에서 유실보다 중복이 안전하고, 중복은 멱등성으로 흡수. 실패 메시지는 DLQ(`platform/kafka`)로.

### 관측성 (운영 권장)

| 영역 | 선택 | 비고 |
|------|------|------|
| 메트릭 | prometheus/client_golang | 배포 큐 적체, 처리 지연, DLQ 적재율 |
| 알림 | Slack (기존 infra/notification 유지) | 배포 실패·DLQ 임계 |
| 테스트 | testing + testify | 외부 의존(pgx/kafka/redis/ethclient)만 mock |

---

## 시작하기

```bash
brew install go golangci-lint
go version   # go1.25 이상

# 의존성 (구현 시점에 추가)
go get entgo.io/ent             # ORM (pgx 드라이버 백엔드)
go get github.com/jackc/pgx/v5
go get github.com/segmentio/kafka-go
go get github.com/redis/go-redis/v9
go get github.com/ethereum/go-ethereum
go get github.com/caarlos0/env/v11
go mod tidy

# Ent 스키마 정의(ent/schema/) 후 코드 생성
make generate

cp .env.example .env          # 값 채우기
export DATABASE_URL="postgres://..."
make migrate-up               # Ent/atlas 마이그레이션

make build && make test
```

> 현재 상태: **골격(skeleton)**. `account`/`deposit` 슬라이스와 platform 패키지의
> 인터페이스·구조체만 정의돼 있고, 나머지 기능 슬라이스는 패키지 선언만 있다.
> 구현부는 `panic("not implemented")` 또는 `TODO(골격)` 표시.
> 새 기능은 `account/`의 5파일 패턴(domain·service·deploy/usecase·store·handler)을 따라 추가한다.

---

## 환경변수

| 변수 | 필수 | 기본값 | 용도 |
|------|------|--------|------|
| `DATABASE_URL` | ✅ | — | Postgres DSN |
| `KAFKA_BROKER` | ✅ | — | Kafka 브로커 |
| `REDIS_URL` | ✅ | — | Redis(분산락) |
| `BUNDLER_URL` | ✅ | — | 외부 번들러 URL |
| `WS_LISTEN` | | :8080 | Listener WS 수신 주소 |
| `DEPLOY_LOCK_TTL_MS` | | 30000 | 배포 분산락 TTL |
| `LOG_LEVEL` | | info | 로그 레벨 |

---

## 관련 프로젝트

| 프로젝트 | 역할 | 연동 |
|----------|------|------|
| StableCoinBC_Adapter_Listener | 입금 감지 | → 이 어댑터로 WS 입금 이벤트 전송 |
| StableCoin_Bundler | ERC-4337 번들러 + CREATE2 배포 실행 | ← 이 어댑터가 HTTP로 호출 |
