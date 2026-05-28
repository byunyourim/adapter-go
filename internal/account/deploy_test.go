package account

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
)

// ── 인터페이스 fake 구현 (외부 의존 없음) ──

type fakeRepo struct {
	mu sync.Mutex
	m  map[string]Account
}

func newFakeRepo() *fakeRepo { return &fakeRepo{m: map[string]Account{}} }

func rkey(chainID int64, addr string) string { return addr } // 테스트 단순화(체인 1개)

func (r *fakeRepo) Get(_ context.Context, chainID int64, address string) (Account, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	a, ok := r.m[rkey(chainID, address)]
	return a, ok, nil
}

func (r *fakeRepo) Save(_ context.Context, a Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.m[rkey(a.ChainID, a.Address)]; ok {
		return nil
	}
	r.m[rkey(a.ChainID, a.Address)] = a
	return nil
}

func (r *fakeRepo) MarkDeployed(_ context.Context, chainID int64, address, txHash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	a := r.m[rkey(chainID, address)]
	a.Deployed = true
	a.DeployTxHash = txHash
	r.m[rkey(chainID, address)] = a
	return nil
}

type fakeLocker struct{ calls int }

func (l *fakeLocker) WithLock(_ context.Context, _ string, fn func() error) error {
	l.calls++
	return fn()
}

type fakeDeployer struct {
	tx    string
	err   error
	calls int
}

func (d *fakeDeployer) Deploy(context.Context, int64, string, string) (string, error) {
	d.calls++
	return d.tx, d.err
}
func (d *fakeDeployer) PredictAddress(context.Context, int64, string) (string, error) {
	return "", nil
}

type fakePublisher struct{ results []DeployResult }

func (p *fakePublisher) PublishDeployResult(_ context.Context, r DeployResult) error {
	p.results = append(p.results, r)
	return nil
}

func newService(repo Repository, locker Locker, dep Deployer, pub ResultPublisher) *Service {
	return NewService(locker, dep, repo, pub, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

// ── 테스트 ──

func TestHandleDeploy_New(t *testing.T) {
	repo := newFakeRepo()
	locker := &fakeLocker{}
	dep := &fakeDeployer{tx: "0xabc"}
	pub := &fakePublisher{}
	svc := newService(repo, locker, dep, pub)

	cmd := DeployCommand{ChainID: 56357, Address: "0xWallet", Salt: "0x01"}
	if err := svc.HandleDeploy(context.Background(), cmd); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if dep.calls != 1 {
		t.Errorf("Deploy 호출 1회 기대, got %d", dep.calls)
	}
	if locker.calls != 1 {
		t.Errorf("락 1회 기대, got %d", locker.calls)
	}
	a, ok, _ := repo.Get(context.Background(), cmd.ChainID, cmd.Address)
	if !ok || !a.Deployed || a.DeployTxHash != "0xabc" {
		t.Errorf("배포 상태 갱신 안 됨: %+v", a)
	}
	if len(pub.results) != 1 || !pub.results[0].Success || pub.results[0].TxHash != "0xabc" {
		t.Errorf("성공 결과 발행 기대, got %+v", pub.results)
	}
}

func TestHandleDeploy_Idempotent(t *testing.T) {
	repo := newFakeRepo()
	// 이미 배포된 계좌 사전 적재.
	_ = repo.Save(context.Background(), Account{ChainID: 56357, Address: "0xWallet", Salt: "0x01"})
	_ = repo.MarkDeployed(context.Background(), 56357, "0xWallet", "0xPREV")

	locker := &fakeLocker{}
	dep := &fakeDeployer{tx: "0xNEW"}
	pub := &fakePublisher{}
	svc := newService(repo, locker, dep, pub)

	cmd := DeployCommand{ChainID: 56357, Address: "0xWallet", Salt: "0x01"}
	if err := svc.HandleDeploy(context.Background(), cmd); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if dep.calls != 0 {
		t.Errorf("이미 배포됨 — Deploy 호출 0회 기대, got %d", dep.calls)
	}
	if locker.calls != 0 {
		t.Errorf("빠른 경로 — 락 0회 기대, got %d", locker.calls)
	}
	if len(pub.results) != 1 || pub.results[0].TxHash != "0xPREV" {
		t.Errorf("기존 txHash 재발행 기대, got %+v", pub.results)
	}
}

func TestHandleDeploy_DeployError(t *testing.T) {
	repo := newFakeRepo()
	locker := &fakeLocker{}
	dep := &fakeDeployer{err: errors.New("rpc timeout")}
	pub := &fakePublisher{}
	svc := newService(repo, locker, dep, pub)

	cmd := DeployCommand{ChainID: 56357, Address: "0xWallet", Salt: "0x01"}
	err := svc.HandleDeploy(context.Background(), cmd)
	if err == nil {
		t.Fatal("배포 실패 시 에러 기대")
	}
	a, _, _ := repo.Get(context.Background(), cmd.ChainID, cmd.Address)
	if a.Deployed {
		t.Error("배포 실패인데 deployed=true")
	}
	if len(pub.results) != 0 {
		t.Errorf("실패 시 성공 결과 발행 금지, got %+v", pub.results)
	}
}
