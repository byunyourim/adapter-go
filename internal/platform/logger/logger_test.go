package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestParseLevel(t *testing.T) {
	cases := map[string]slog.Level{
		"trace": slog.LevelDebug,
		"debug": slog.LevelDebug,
		"INFO":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
		"fatal": slog.LevelError,
		"":      slog.LevelInfo,
		"bogus": slog.LevelInfo,
	}
	for in, want := range cases {
		if got := ParseLevel(in); got != want {
			t.Errorf("ParseLevel(%q)=%v, want %v", in, got, want)
		}
	}
}

func TestPinoLabelMapping(t *testing.T) {
	cases := map[slog.Level]string{
		slog.LevelDebug: "debug",
		slog.LevelInfo:  "info",
		slog.LevelWarn:  "warn",
		slog.LevelError: "error",
	}
	for lvl, want := range cases {
		if got := pinoLabel(lvl); got != want {
			t.Errorf("pinoLabel(%v)=%q, want %q", lvl, got, want)
		}
	}
}

// TS 어댑터 pino 포맷 검증: level은 문자열 라벨, time은 KST ISO(+09:00), module 포함.
func TestJSONOutputMatchesTSAdapter(t *testing.T) {
	var buf bytes.Buffer
	log := build(&buf, "adapter", slog.LevelInfo, false)
	log.Info("hello", "chain", 56357)

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("출력이 JSON이 아님: %v\n%s", err, buf.String())
	}

	if m["level"] != "info" {
		t.Errorf("level=%v, want \"info\"(string)", m["level"])
	}

	ts, ok := m["time"].(string)
	if !ok {
		t.Fatalf("time이 문자열이 아님: %v", m["time"])
	}
	if !strings.HasSuffix(ts, "+09:00") {
		t.Errorf("time이 KST(+09:00) 아님: %q", ts)
	}
	if _, err := time.Parse(kstLayout, ts); err != nil {
		t.Errorf("time이 KST ISO 레이아웃과 안 맞음: %q (%v)", ts, err)
	}

	if m["msg"] != "hello" {
		t.Errorf("msg=%v, want hello", m["msg"])
	}
	if m["module"] != "adapter" {
		t.Errorf("module=%v, want adapter", m["module"])
	}
	if _, ok := m["hostname"]; !ok {
		t.Error("hostname 필드 없음")
	}
	if _, ok := m["pid"]; !ok {
		t.Error("pid 필드 없음")
	}
}
