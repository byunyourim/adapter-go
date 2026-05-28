// Package bundler 는 외부 번들러(StableCoin_Bundler) HTTP 클라이언트다.
// (TS의 adapter/out/bundler/external-bundler 대응) ERC-4337 UserOperation 제출.
package bundler

import "context"

// Client 는 번들러 API 클라이언트.
type Client struct {
	baseURL string
	// TODO(골격): *http.Client
}

// New 는 클라이언트를 만든다.
func New(baseURL string) *Client {
	return &Client{baseURL: baseURL}
}

// SendUserOp 는 UserOperation을 번들러에 제출한다.
//
// TODO(골격): net/http POST + 응답/에러 매핑.
func (c *Client) SendUserOp(ctx context.Context, chainID int64, userOp []byte) (string, error) {
	panic("not implemented")
}
