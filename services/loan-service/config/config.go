package config

import (
	"books-and-trust/shared/env"
)

type Config struct {
	Database dbConfig
	App      appConfig
	Tracing  tracingConfig
}

type dbConfig struct {
	Addr         string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  string
}

type appConfig struct {
	Addr string
	Environment string
	ServiceName string

}

type tracingConfig struct {
	JaegerEndpoint string
}
func LoadConfigs() *Config {
	return &Config{
		Database: dbConfig{
			Addr:         env.GetString("LOANS_DB_URI", ""),
			MaxOpenConns: env.GetInt("LOANS_DB_MAX_OPEN_CONNS", 30),
			MaxIdleConns: env.GetInt("LOANS_DB_MAX_IDLE_CONNS", 30),
			MaxIdleTime:  env.GetString("LOANS_DB_MAX_IDLE_TIME", "15m"),
		},
		App: appConfig{
			Addr: env.GetString("LOAN_SERVICE_ADDR", ":9092"),
			Environment: env.GetString("ENVIRONMENT", "development"),
			ServiceName: "loan-service",
		},
		Tracing: tracingConfig{
			JaegerEndpoint: env.GetString("JAEGER_ENDPOINT", "jaeger:4317"),
		},
	}
}
