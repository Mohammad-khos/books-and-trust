package config

import "books-and-trust/shared/env"

type Config struct {
	App   appConfig
	Grpc  grpcClientConfig
	Redis redisConfig
	Trace tracingConfig
}

type appConfig struct {
	Addr string
	ServiceName string
}

type grpcClientConfig struct {
	UserClient userClientConfig
	LoanClient loanClientConfig
}

type userClientConfig struct {
	Addr string
}
type loanClientConfig struct {
	Addr string
}

type redisConfig struct {
	Addr     string
	Password string
	DB       int
}

type tracingConfig struct {
	Environment    string
	JaegerEndpoint string
}

func LoadConfigs() *Config {
	return &Config{
		App: appConfig{
			Addr: env.GetString("APIGATEWAY_ADDR" , ":8081"),
			ServiceName: "api-gateway",
		},
		Grpc: grpcClientConfig{
			UserClient: userClientConfig{
				Addr: "user_service" + env.GetString("USER_SERVICE_ADDR" , ":9091"),
			},
			LoanClient: loanClientConfig{
				Addr: "loan_service" + env.GetString("LOAN_SERVICE_ADDR" , ":9092"),
			},
		},
		Redis: redisConfig{
			Addr: env.GetString("REDIS_ADDR" , "redis:6379"),
			Password: env.GetString("REDIS_PASSWORD" , ""),
			DB: env.GetInt("REDIS_DB" , 0),
		},
		Trace: tracingConfig{
			Environment: env.GetString("ENVIRONMENT" , "development"),
			JaegerEndpoint: env.GetString("JAEGER_ENDPOINT", "jaeger:4317"),
		},
	}
}
