// Package logger 구조화 로깅 제공 (TS의 infra/logger = pino 대응)
//
// 운영(ELK)은 pino 포맷 호환 JSON, LOG_PRETTY면 text 핸들러로 전환
package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
	"time"
)

// kstLayout TS 어댑터 timestamp와 동일한 KST ISO 포맷
const kstLayout = "2006-01-02T15:04:05.000-07:00"

// kst KST 고정 오프셋(+09:00, DST 없음 — tzdata 의존 회피)
var kst = time.FixedZone("KST", 9*60*60)

var hostname = resolveHostname()

// New TS 어댑터 pino 포맷과 동일한 구조화 로거 생성
// level/pretty는 config(env)에서 받아 주입한다(process.env 직접 접근 금지 원칙)
func New(module string, level slog.Level, pretty bool) *slog.Logger {
	return build(os.Stdout, module, level, pretty)
}

// build writer를 받아 로거 생성(테스트에서 버퍼 주입용)
func build(w io.Writer, module string, level slog.Level, pretty bool) *slog.Logger {
	var h slog.Handler
	if pretty {
		h = slog.NewTextHandler(w, &slog.HandlerOptions{Level: level})
	} else {
		h = slog.NewJSONHandler(w, &slog.HandlerOptions{
			Level:       level,
			ReplaceAttr: pinoReplace,
		})
	}
	return slog.New(h).With(
		slog.String("module", module),
		slog.Int("pid", os.Getpid()),
		slog.String("hostname", hostname),
	)
}

// ParseLevel LOG_LEVEL 문자열을 slog.Level로 변환(기본 info)
// pino의 trace/fatal은 slog 대응 레벨이 없어 debug/error로 매핑
func ParseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "trace", "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error", "fatal":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// pinoReplace time/level 필드를 TS 어댑터(pino) 포맷으로 치환
func pinoReplace(_ []string, a slog.Attr) slog.Attr {
	switch a.Key {
	case slog.TimeKey:
		return slog.String(slog.TimeKey, a.Value.Time().In(kst).Format(kstLayout))
	case slog.LevelKey:
		lvl, _ := a.Value.Any().(slog.Level)
		return slog.String(slog.LevelKey, pinoLabel(lvl))
	default:
		return a
	}
}

// pinoLabel slog 레벨을 pino 문자열 라벨로 변환
func pinoLabel(l slog.Level) string {
	switch {
	case l >= slog.LevelError:
		return "error"
	case l >= slog.LevelWarn:
		return "warn"
	case l >= slog.LevelInfo:
		return "info"
	case l >= slog.LevelDebug:
		return "debug"
	default:
		return "trace"
	}
}

func resolveHostname() string {
	h, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return h
}
