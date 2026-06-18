package main

import (
	"books-and-trust/services/api-gateway/config"
	"books-and-trust/services/api-gateway/internal/client"
	handler "books-and-trust/services/api-gateway/internal/handler/http"
	loanHandler "books-and-trust/services/api-gateway/internal/handler/http/loan_handler"
	userHandler "books-and-trust/services/api-gateway/internal/handler/http/user_handler"
	"books-and-trust/services/api-gateway/internal/infra/cache"
	circuitbreaker "books-and-trust/services/api-gateway/internal/infra/circuit-breaker"
	"books-and-trust/services/api-gateway/internal/infra/metrics"
	"books-and-trust/services/api-gateway/internal/middleware"
	"books-and-trust/services/api-gateway/internal/ratelimiter"
	"books-and-trust/shared/log"
	"books-and-trust/shared/retry"

	p "books-and-trust/shared/metrics/prometheus"
	"books-and-trust/shared/tracing"
	"context"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"go.uber.org/zap"
)

type Application struct {
	Config     *config.Config
	Router     *chi.Mux
	Logger     *zap.SugaredLogger
	Handler    *handler.HTTPHandler
	Middleware middleware.GatewayMiddleware
	metrics    *metrics.Metrics
	tracer     *tracing.Tracer
}

var reg *prometheus.Registry = prometheus.NewRegistry()

func NewApplication() (*Application, func()) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	//logger
	logger, loggerSync := log.NewLogger()
	//config
	cfg := config.LoadConfigs()
	//router
	mux := chi.NewMux()
	//circuit breaker
	loanCB := circuitbreaker.NewLoanServiceBreaker()
	//user grpc client
	userClient, err := client.NewUserClient(cfg.Grpc.UserClient.Addr)
	if err != nil {
		logger.Fatalw("error connecting user grpc client", "error", err)
	}
	//loan grpc client
	loanClient, err := client.NewLoanClient(cfg.Grpc.LoanClient.Addr, loanCB)
	if err != nil {
		logger.Fatalw("error connecting loan grpc client", "error", err)
	}
	//user handler
	userHandler := userHandler.NewUserHandler(logger, userClient)
	//loan handler
	loanHandler := loanHandler.NewLoanHandler(logger, loanClient)
	//general handler
	handler := handler.NewHTTPHandler(logger, userHandler, loanHandler)
	//tracer
	tracer, err := tracing.NewTracer(cfg.Trace.JaegerEndpoint, cfg.App.ServiceName, cfg.Trace.Environment)
	if err != nil {
		logger.Errorw("error initializing tracer", "error", err)
	}
	//cache strorage
	var cacheStore cache.CacheStore
	err = retry.Do(ctx, retry.Config{
		Attempts: 5,
		Initial:  1 * time.Second,
		MaxDelay: 5 * time.Second,
		Factor:   2,
	}, func() error {
		var err error
		cacheStore, err = cache.NewRedisStore(
			cfg.Redis.Addr,
			cfg.Redis.Password,
			cfg.Redis.DB,
		)
		return err
	}, nil)

	if err != nil {
		logger.Errorw("error creating redis cache storage connection", "error", err)
	}

	//rate limiter
	rt, err := ratelimiter.NewTokenBucketLimiter(cacheStore)
	if err != nil {
		logger.Errorw("error creating token bucket rate limiter", "error", err)
	}
	//metrics
	factory := p.NewFactory("myapp", "gateway")
	metrics := metrics.NewMetrics(factory)
	reg.MustRegister(
		metrics.RequestTotal,
		metrics.RequestLatency,
		metrics.InFlight,
		collectors.NewProcessCollector(
			collectors.ProcessCollectorOpts{},
		),
		collectors.NewGoCollector(),
	)
	// middlewares
	authMidd := middleware.NewAuthMiddleware(logger, userClient)
	cspMidd := middleware.NewCSPMIddleware("self", "self", "self")
	adminMidd := middleware.NewAdminMiddleware("ADMINS_FILE_PATH", logger)
	limiter := middleware.NewRateLimiterMiddleware(rt, logger, 1000, 10)
	securityMidd := middleware.NewSecurityHeadersMiddleware()
	reqMidd := middleware.NewRequestLoggerMiddleware(logger)
	metricsMidd := middleware.NewMetricsMiddleware(metrics)
	mw := middleware.NewGatewayMiddleware(authMidd, cspMidd, adminMidd, limiter, securityMidd, reqMidd, metricsMidd)
	cleanup := func() {
		logger.Infow("closing all downstream grpc connections...")
		userClient.Close()
		loanClient.Close()
		loggerSync()
		cacheStore.Close()
		tracer.Shutdown(ctx)
	}

	return &Application{
		Logger:     logger,
		Router:     mux,
		Handler:    handler,
		Config:     cfg,
		Middleware: mw,
		metrics:    metrics,
		tracer:     tracer,
	}, cleanup
}
