// Package env 는 환경변수 파싱을 전담한다. (TS의 infra/config/env 대응)
// process.env 직접 접근은 이 패키지에서만.
package env

import "time"

// Config 는 어댑터 전역 설정.
type Config struct {
	DatabaseURL string `env:"DATABASE_URL,required"`
	KafkaBroker string `env:"KAFKA_BROKER,required"`
	RedisURL    string `env:"REDIS_URL,required"`
	BundlerURL  string `env:"BUNDLER_URL,required"`
	WSListen    string `env:"WS_LISTEN" envDefault:":8080"` // Listener WS 수신 주소

	DeployLockTTL time.Duration `env:"DEPLOY_LOCK_TTL_MS" envDefault:"30s"`
	LogLevel      string        `env:"LOG_LEVEL" envDefault:"info"`
}

// Load 는 환경변수를 파싱한다. 필수 누락 시 error.
//
// TODO(골격): caarlos0/env 또는 viper로 구현. 새 env 추가 시 .env.example, README 동기화.
func Load() (*Config, error) {
	panic("not implemented")
}
