// Package apperr 는 어댑터 공통 에러 체계다. (TS의 domain/error/* 대응)
// 에러는 값으로 다루고 %w로 래핑해 전파, 분류는 errors.Is/As로 판단한다.
package apperr

import "errors"

// 분류용 sentinel 에러.
var (
	ErrNotFound   = errors.New("not found")
	ErrValidation = errors.New("validation failed")
	ErrConflict   = errors.New("conflict")        // 멱등 충돌 등
	ErrRetryable  = errors.New("retryable")        // 재시도 가능 인프라 오류
)

// BusinessError 는 도메인 규칙 위반. 부가 코드를 담는다.
type BusinessError struct {
	Code string
	Msg  string
}

func (e *BusinessError) Error() string { return e.Code + ": " + e.Msg }

// IsRetryable 은 에러가 재시도 대상인지 판단한다.
//
// TODO(골격): RPC/네트워크/Kafka 에러 분류 (listener retry.IsRetryable 참고).
func IsRetryable(err error) bool {
	return errors.Is(err, ErrRetryable)
}
