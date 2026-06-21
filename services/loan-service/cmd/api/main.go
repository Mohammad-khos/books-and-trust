package main

import (
	"books-and-trust/services/loan-service/config"
	handler "books-and-trust/services/loan-service/internal/handler/grpc"
	"books-and-trust/services/loan-service/internal/infra/repo"
	"books-and-trust/services/loan-service/internal/service"
	"books-and-trust/shared/db"
	"books-and-trust/shared/log"
	pb "books-and-trust/shared/proto/loan"
	"books-and-trust/shared/tracing"
	"context"
	"errors"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
)

func main() {
	//creating context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	traceCtx, traceCancel := context.WithTimeout(ctx, time.Second*10)
	defer traceCancel()
	//configs
	cfg := config.LoadConfigs()
	//logger
	logger, loggerSync := log.NewLogger()
	defer loggerSync()
	//db
	db, err := db.New(cfg.Database.Addr, cfg.Database.MaxOpenConns, cfg.Database.MaxIdleConns, cfg.Database.MaxIdleTime)
	if err != nil {
		logger.Panicw("failed to connect postgres database", "error", err)
	}
		//closing db connection at the end
	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatalw("failed to get sql db", "error", err)
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			logger.Errorw("failed to close Sql database", "error", err)
		}
	}()

	logger.Info("database successfully established")

	//repository
	repo := repo.NewPostgresRepository(db)
	//service
	svc := service.NewLoanService(repo)
	//trace
	tracer, err := tracing.NewTracer(cfg.Tracing.JaegerEndpoint, cfg.App.ServiceName, cfg.App.Environment)
	if err != nil {
		logger.Errorw("error initializing tracer", "error", err)
	}
	defer func() {
		if err := tracer.Shutdown(traceCtx); err != nil {
			logger.Errorw("error shutting down tracer", "error", err)
		}
	}()
	//creating tcp connection
	listener, err := net.Listen("tcp", cfg.App.Addr)
	if err != nil {
		logger.Panicw("failed to create tcp connection", "error", err)
	}
	grpcServer := grpc.NewServer(tracing.WithTracingInterceptors()...)
	handler := handler.NewGRPCHandler(svc)
	pb.RegisterLoanServiceServer(grpcServer, handler)
	//os signal goroutin
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		defer signal.Stop(sig)
		<-sig
		cancel()
	}()
	//grpc server goroutin
	go func() {
		logger.Info("grpc server has started successfully")
		if err := grpcServer.Serve(listener); err != nil &&
			!errors.Is(err, grpc.ErrServerStopped) {

			logger.Errorw("failed to serve grpc", "error", err)
			cancel()
		}
	}()
	//gracefull shutdown
	<-ctx.Done()
	logger.Info("grpc server is shutting down")
	grpcServer.GracefulStop()
}
