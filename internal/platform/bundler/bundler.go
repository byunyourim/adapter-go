// Package bundler 외부 번들러(StableCoin_Bundler) HTTP 클라이언트
// (TS의 adapter/out/bundler/external-bundler 대응) ERC-4337 UserOperation 제출
package bundler

import "context"

// Client 번들러 API 클라이언트
type Client struct {
	baseURL string
	// TODO(골격): *http.Client
}

// New 클라이언트 생성
func New(baseURL string) *Client {
	return &Client{baseURL: baseURL}
}

// SendUserOp UserOperation을 번들러에 제출
//
// TODO(골격): net/http POST 구현. 번들러는 실패 시 tx-error.ts로 정밀 분류한
// 구조화 에러를 응답한다(HTTP: {error,code,category,retryable,txHash},
// JSON-RPC: error.data.{code,category,retryable}). 그 필드를 ClassifyResponse로 넘긴다:
//
//	resp, status, err := c.post(ctx, endpoint, userOp)
//	if err != nil { return "", bundler.Classify(err) }           // 전송 자체 실패(네트워크)
//	if status >= 400 {
//		var b struct{ Code string; Retryable bool; Error string }
//		_ = json.Unmarshal(resp, &b)
//		return "", bundler.ClassifyResponse(b.Code, b.Retryable, errors.New(b.Error))
//	}
func (c *Client) SendUserOp(ctx context.Context, chainID int64, userOp []byte) (string, error) {
	panic("not implemented")
}
