package account

import (
	"context"

	"github.com/byunyourim/stablecoinbc-adapter/ent"
	entaccount "github.com/byunyourim/stablecoinbc-adapter/ent/account"
)

// Store account 영속 구현(Ent). Repository 인터페이스를 만족
// 돈이 걸린 핵심 경로는 Ent 트랜잭션(client.Tx) 또는 raw SQL로 명시 제어
type Store struct {
	client *ent.Client
}

// NewStore Store 생성
func NewStore(client *ent.Client) *Store {
	return &Store{client: client}
}

// Get 계좌 조회. 없으면 (_, false, nil)
func (s *Store) Get(ctx context.Context, chainID int64, address string) (Account, bool, error) {
	row, err := s.client.Account.Query().
		Where(entaccount.ChainID(chainID), entaccount.Address(address)).
		Only(ctx)
	if ent.IsNotFound(err) {
		return Account{}, false, nil
	}
	if err != nil {
		return Account{}, false, err
	}
	return toDomain(row), true, nil
}

// Save 계좌 저장(이미 있으면 무시 — 멱등)
// deploy.go의 체인 단위 락 안에서 호출되므로 check-then-create가 안전하며,
// 최종 방어는 (chain_id, address) 유니크 인덱스다
func (s *Store) Save(ctx context.Context, a Account) error {
	if _, ok, err := s.Get(ctx, a.ChainID, a.Address); err != nil {
		return err
	} else if ok {
		return nil
	}
	return s.client.Account.Create().
		SetChainID(a.ChainID).
		SetAddress(a.Address).
		SetSalt(a.Salt).
		SetDeployed(false).
		Exec(ctx)
}

// MarkDeployed 배포 완료로 표시. 동일 값 재적용도 무해(멱등)
func (s *Store) MarkDeployed(ctx context.Context, chainID int64, address, txHash string) error {
	_, err := s.client.Account.Update().
		Where(entaccount.ChainID(chainID), entaccount.Address(address)).
		SetDeployed(true).
		SetDeployTxHash(txHash).
		Save(ctx)
	return err
}

func toDomain(row *ent.Account) Account {
	return Account{
		ChainID:      row.ChainID,
		Address:      row.Address,
		Salt:         row.Salt,
		Deployed:     row.Deployed,
		DeployTxHash: row.DeployTxHash,
	}
}
