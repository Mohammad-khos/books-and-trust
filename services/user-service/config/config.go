package config

import (
	"books-and-trust/shared/env"
	"time"
)

type Config struct {
	Database dbConfig
	App      appConfig
	Jwt      jwtConfig
	Tracer   tracingConfig
}

type appConfig struct {
	Addr        string
	ServiceName string
	Environment    string
}

type dbConfig struct {
	Addr         string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  string
}

type jwtConfig struct {
	Secret string
	Aud    string
	Iss    string
	Exp    time.Duration
}
type tracingConfig struct {
	JaegerEndpoint string
}

func LoadConfigs() *Config {
	return &Config{
		Database: dbConfig{
			Addr:         env.GetString("USERS_DB_URI", ""),
			MaxOpenConns: env.GetInt("USERS_DB_MAX_OPEN_CONNS", 30),
			MaxIdleConns: env.GetInt("USERS_DB_MAX_IDLE_CONNS", 30),
			MaxIdleTime:  env.GetString("USERS_DB_MAX_IDLE_TIME", "15m"),
		},
		App: appConfig{
			Addr: env.GetString("USER_SERVICE_ADDR", ":9091"),
			ServiceName: "user-service",
			Environment:    env.GetString("ENVIRONMENT" , "development"),
		},
		Jwt: jwtConfig{
			Secret: env.GetString("JWT_SECRET", "jwt_secret"),
			Aud:    env.GetString("JWT_AUD" , "books-and-trust"),
			Iss:    env.GetString("JWT_AUD" , "books-and-trust"),
			Exp:    time.Hour * 24 * 7,
		},
		Tracer: tracingConfig{
			JaegerEndpoint: env.GetString("JAEGER_ENDPOINT", "jaeger:4317"),
		},
	}
}
