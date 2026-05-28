// Package logger 는 구조화 로깅 팩토리다. (TS의 infra/logger 대응) log/slog 사용.
package logger

import (
	"log/slog"
	"os"
)

// New 는 name 필드가 붙은 JSON 구조화 로거를 만든다.
//
// TODO(골격): LOG_LEVEL 반영, request-context의 trace id 부착.
func New(name string) *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("logger", name)
}
