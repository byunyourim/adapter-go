package account

import (
	"context"
	"log/slog"
)

// 소비자 측 인터페이스 — 이 패키지가 필요한 만큼만 좁게 선언한다(Go 관용).
// 구현체는 platform/* 패키지가 제공하며, wiring 시점에 주입된다.

// Locker 는 디플로이어 계정 단위 분산락(Redis). nonce 직렬화 격리에 사용.
type Locker interface {
	WithLock(ctx context.Context, key string, fn func() error) error
}

// Deployer 는 CREATE2 지갑 배포 트랜잭션을 제출하고 확정까지 대기한다.
type Deployer interface {
	Deploy(ctx context.Context, chainID int64, salt, address string) (txHash string, err error)
	PredictAddress(ctx context.Context, chainID int64, salt string) (string, error)
}

// Repository 는 계좌 영속(Postgres/Ent).
type Repository interface {
	// Get 은 계좌를 조회한다. 없으면 (_, false, nil).
	Get(ctx context.Context, chainID int64, address string) (Account, bool, error)
	// Save 는 계좌를 저장한다(이미 있으면 무시 — 멱등).
	Save(ctx context.Context, a Account) error
	// MarkDeployed 는 배포 완료로 표시한다.
	MarkDeployed(ctx context.Context, chainID int64, address, txHash string) error
}

// ResultPublisher 는 처리 결과를 Kafka로 발행.
type ResultPublisher interface {
	PublishDeployResult(ctx context.Context, r DeployResult) error
}

// Service 는 account 기능의 유스케이스를 묶는다.
type Service struct {
	locker  Locker
	deploy  Deployer
	repo    Repository
	results ResultPublisher
	log     *slog.Logger
}

// NewService 는 의존성을 주입받아 Service를 만든다(수동 DI).
func NewService(locker Locker, deploy Deployer, repo Repository, results ResultPublisher, log *slog.Logger) *Service {
	return &Service{locker: locker, deploy: deploy, repo: repo, results: results, log: log}
}
