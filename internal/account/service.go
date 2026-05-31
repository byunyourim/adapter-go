package account

import (
	"context"
	"log/slog"
)

// 소비자 측 인터페이스 — 이 패키지가 필요한 만큼만 좁게 선언(Go 관용)
// 구현체는 platform/* 패키지가 제공, wiring 시점에 주입

// Locker 디플로이어 계정 단위 분산락(Redis). nonce 직렬화 격리에 사용
type Locker interface {
	WithLock(ctx context.Context, key string, fn func() error) error
}

// Deployer CREATE2 지갑 배포 트랜잭션을 제출하고 확정까지 대기
//
// Deploy 는 멱등이어야 함: 해당 주소에 이미 컨트랙트 코드가 있으면 재전송 없이
// 기존(또는 빈) txHash를 반환. DB 상태 갱신 실패 후 재처리되는 at-least-once
// 경로에서 중복 배포(=revert) 트랜잭션을 막기 위함
type Deployer interface {
	Deploy(ctx context.Context, chainID int64, salt, address string) (txHash string, err error)
	PredictAddress(ctx context.Context, chainID int64, salt string) (string, error)
}

// Repository 계좌 영속(Postgres/Ent)
type Repository interface {
	// Get 계좌 조회. 없으면 (_, false, nil)
	Get(ctx context.Context, chainID int64, address string) (Account, bool, error)
	// Save 계좌 저장(이미 있으면 무시 — 멱등)
	Save(ctx context.Context, a Account) error
	// MarkDeployed 배포 완료로 표시
	MarkDeployed(ctx context.Context, chainID int64, address, txHash string) error
}

// ResultPublisher 처리 결과를 Kafka로 발행
type ResultPublisher interface {
	PublishDeployResult(ctx context.Context, r DeployResult) error
}

// Service account 기능의 유스케이스를 묶음
type Service struct {
	locker  Locker
	deploy  Deployer
	repo    Repository
	results ResultPublisher
	log     *slog.Logger
}

// NewService 의존성을 주입받아 Service 생성(수동 DI)
func NewService(locker Locker, deploy Deployer, repo Repository, results ResultPublisher, log *slog.Logger) *Service {
	return &Service{locker: locker, deploy: deploy, repo: repo, results: results, log: log}
}
