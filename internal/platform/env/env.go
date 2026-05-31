// Package env 환경변수 파싱 전담 (TS의 infra/config/env 대응)
// process.env 직접 접근은 이 패키지에서만
package env

import "time"

// Config 어댑터 전역 설정
type Config struct {
	DatabaseURL string `env:"DATABASE_URL,required"`
	KafkaBroker string `env:"KAFKA_BROKER,required"`
	RedisURL    string `env:"REDIS_URL,required"`
	BundlerURL  string `env:"BUNDLER_URL,required"`
	WSListen    string `env:"WS_LISTEN" envDefault:":8080"` // Listener WS 수신 주소

	DeployLockTTL time.Duration `env:"DEPLOY_LOCK_TTL_MS" envDefault:"30s"`

	// 로깅 — TS 어댑터와 동일 기본값. 운영(ELK)에서는 LOG_PRETTY=false로 JSON 출력
	LogLevel  string `env:"LOG_LEVEL" envDefault:"debug"`
	LogPretty bool   `env:"LOG_PRETTY" envDefault:"true"`
}

// Load 환경변수 파싱 — 필수 누락 시 error
//
// TODO(골격): caarlos0/env 또는 viper로 구현. 새 env 추가 시 .env.example, README 동기화
func Load() (*Config, error) {
	panic("not implemented")
}
