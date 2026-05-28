// Package blockchain 은 go-ethereum 기반 체인 연동을 제공한다.
// (TS의 adapter/out/blockchain/ethers-* 대응) ethclient + abigen 생성 바인딩.
package blockchain

import (
	"context"
	"math/big"
)

// Client 는 체인 읽기 + 트랜잭션 전송을 묶는다.
type Client struct {
	// TODO(골격): *ethclient.Client, 디플로이어 키(KMS 권장), nonce 관리자
}

// New 는 RPC URL로 클라이언트를 연다.
//
// TODO(골격): ethclient.Dial.
func New(ctx context.Context, rpcURL string) (*Client, error) {
	panic("not implemented")
}

// Deploy 는 CREATE2 지갑 배포 트랜잭션을 제출하고 확정까지 대기한다.
// account.Deployer 인터페이스를 만족한다.
//
// nonce 직렬화 주의: 디플로이어 계정의 nonce는 여기서 원자적으로 관리해야 한다
// (Redis 기반 nonce allocator 또는 단일 직렬 실행). README 참고.
func (c *Client) Deploy(ctx context.Context, chainID int64, salt, address string) (string, error) {
	panic("not implemented")
}

// PredictAddress 는 배포 전 CREATE2 입금 주소를 계산한다.
func (c *Client) PredictAddress(ctx context.Context, chainID int64, salt string) (string, error) {
	panic("not implemented")
}

// BalanceOf 는 잔액을 조회한다.
func (c *Client) BalanceOf(ctx context.Context, chainID int64, address, token string) (*big.Int, error) {
	panic("not implemented")
}
