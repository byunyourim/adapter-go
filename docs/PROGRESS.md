# 작업 현황 (PROGRESS)

> 이 문서는 "어디까지 했고, 다음에 뭘 하면 되는지"를 정리한다. 설계 배경은 `README.md`,
> 작업 규칙은 `CLAUDE.md` 참고. 최종 갱신: 2026-05-31

---

## 한 줄 요약

**골격(skeleton) 단계.** Kafka 토픽·메시지 계약(전 슬라이스)과 에러/로깅/번들러-에러 분류,
`account` 배포 로직 + ent 코드 생성까지 완료. **실제 비즈니스 로직과 인프라 클라이언트
구현, Kafka 송수신 동작, main wiring 은 아직 미구현.**

---

## ✅ 완료된 작업

### 메시지 계약 (이번 작업)
- **Kafka 토픽 상수 + 페이로드 구조체를 전 슬라이스에 정의** — `go doc`/`pkgsite`로 문서화
  - 설계서 부록 B 기준 **15개**(신뢰): `account` / `deposit` / `confirm` / `withdraw` / `payment`
  - 컨벤션 **추론 5개**(⚠️ `TODO(설계)`): `balance` / `settlement` / `settlementcollect` / `config` / `collection`
- 페이로드 json 태그 **snake_case 통일** (와이어=snake_case, Go 필드=MixedCaps, `encoding/json` 자동 매핑)
- README에 "문서 보기(godoc/pkgsite)" 섹션 + 토픽 계약 위치 표 추가

### 구현 + 테스트 있음
- `account` 슬라이스 — 배포(CREATE2) 로직: `service.go` / `deploy.go` / `store.go` / `handler.go` (`deploy_test.go` 통과)
- `platform/apperr` — 에러 코드 체계 (테스트 있음)
- `platform/bundler` — 번들러 에러 분류 (테스트 있음)
- `platform/logger` — pino 호환 구조화 로깅(JSON, ELK용) (테스트 있음)
- `platform/kafka` — Producer/Consumer 구현 (kafka-go, at-least-once: 핸들러 성공 후 commit, 실패 시 DLQ, Murmur2 파티셔너로 kafkajs 호환) (테스트 있음)
- `ent/` — ORM 코드 생성 완료 (`ent/schema/account.go` → 클라이언트·쿼리)

---

## ⏳ 미구현 (골격만 있음 — `panic("not implemented")` / `TODO`)

### 인프라 (platform/) — **다른 모든 작업의 전제**
- `platform/postgres` — 풀 생성
- `platform/redis` — 분산락(nonce 격리)
- `platform/blockchain` — go-ethereum RPC 클라이언트
- `platform/bundler` — 번들러 HTTP 클라이언트 (에러 분류만 됨)
- `platform/env` — 환경변수 로딩

### 도메인 슬라이스 (계약만 있고 서비스 로직 없음)
- `deposit` — `Handle` 가 `panic` (멱등 확인 → 적재 → 미배포면 배포 요청)
- `withdraw` / `payment` / `confirm` — 메시지 구조체만, 서비스·핸들러 없음
- `balance` / `settlement` / `settlementcollect` / `config` / `collection` — 추론 계약만

### 진입점
- `cmd/adapter/main.go` — wiring 미구현 (DI, Kafka consumer 라우팅 등록)

---

## 📋 다음 작업 (우선순위)

1. ~~`platform/kafka` 구현~~ ✅ 완료 (Producer/Consumer, at-least-once + DLQ, 테스트).
   남은 보강: 일시적 실패 재시도(현재는 핸들러 내부 책임), DLQ 토픽 사전 생성, ClientID(KAFKA_CLIENT_ID) 적용.
2. **인프라 클라이언트 구현** — `postgres` / `redis` / `blockchain` / `bundler` / `env`. (다음 최우선)
3. **`cmd/adapter/main.go` wiring** — 의존성 주입 + Kafka consumer에 슬라이스별 핸들러(토픽 상수) 라우팅 등록.
4. **도메인 슬라이스 서비스 구현** (`account` 5파일 패턴 따라). 설계서가 명확한 순서로:
   `deposit` → `withdraw` → `payment` → `confirm`.
5. **추론 5개 슬라이스 스펙 확정** — WalletBE와 토픽명·페이로드 확정 후 `balance`/`settlement`/`settlementcollect`/`config`/`collection` 수정. 특히 `settlement`/`settlementcollect`는 핵심 필드가 비어 있음.
6. **입금 누락 방지 불변식 테스트** — 커서 전진 / at-least-once / 멱등성 (CLAUDE.md 요구).

---

## ⚠️ 알려진 이슈 / 확정 필요

- **추론 5개 토픽/페이로드는 근거 없음** — 설계서·TS 어디에도 없어 컨벤션으로 지어낸 초안. 각 파일 `TODO(설계)` 참고. 토픽명(`settlement.collect` vs `settlement_collect` 등)도 미확정.
- **`withdraw.status` / `withdraw.confirmed` / `payment.result`** — 설계서에 페이로드 샘플이 없어 유사 토픽에서 추론(`TODO(설계)`).
- **env 변수 불일치** — README 표는 `KAFKA_BROKER`(단수)인데 설계서는 `KAFKA_BROKERS`(복수) + `KAFKA_CLIENT_ID` + `KAFKA_GROUP_ID`. `config`/`platform/env` 구현 시 정리 필요.
- **`make generate` 워크스페이스 충돌** — `-mod=mod`가 go.work와 충돌. `GOWORK=off go generate ./...` 사용.

---

## 🔭 선택 사항 (필요 시)

- **AsyncAPI 스펙(`asyncapi.yaml`)** — WalletBE와 공유하는 정식 계약·검증·버전 관리가 필요해지면, 현재 Go 구조체를 소스로 작성. (godoc는 내부 개발용 레퍼런스)
