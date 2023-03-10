package app

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/leporo/sqlf"
	config "github.com/vitermakov/otusgo-final/internal/app/config/brutefp"
	deps "github.com/vitermakov/otusgo-final/internal/app/deps/brutefp"
	"github.com/vitermakov/otusgo-final/internal/handler/grpc"
	"github.com/vitermakov/otusgo-final/internal/ratelimit"
	"github.com/vitermakov/otusgo-final/pkg/logger"
	"github.com/vitermakov/otusgo-final/pkg/utils/closer"
	"github.com/vitermakov/otusgo-final/pkg/utils/pgconn"
)

type BruteFP struct {
	config   config.Config
	logger   logger.Logger
	deps     *deps.Deps
	services *deps.Services
	closer   *closer.Closer
}

func NewBruteFP(ctx context.Context, config config.Config) (App, error) {
	logLevel, err := logger.ParseLevel(config.Logger.Level)
	if err != nil {
		return nil, fmt.Errorf("'%s': %w", config.Logger.Level, err)
	}
	logs, err := logger.NewLogrus(logger.Config{
		Level:    logLevel,
		FileName: config.Logger.FileName,
	})
	if err != nil {
		return nil, fmt.Errorf("unable start logger: %w", err)
	}

	closes := closer.NewCloser()
	var dbPool *sql.DB
	if config.Storage.Type == deps.StoreTypeInPgsql {
		pool, closeFn := pgconn.NewPgConn(config.ServiceID, config.Storage.PGConn, logs)
		if pool == nil {
			return nil, fmt.Errorf("unable start logger: %w", err)
		}
		dbPool = pool
		closes.Register("DB", closeFn)

		// устанавливаем диалект билдера запросов
		sqlf.SetDialect(sqlf.PostgreSQL)
		// это костыль, так как при большом количестве запросов он подтекает
		go func() {
			for {
				sqlf.PostgreSQL.ClearCache()
				sqlf.NoDialect.ClearCache()
				select {
				case <-ctx.Done():
					return
				case <-time.After(30 * time.Minute):
				}
			}
		}()
	}
	rateLimiter, closeFn, err := ratelimit.NewRateLimiter(config.Limits)
	closes.Register("Rate Limiter", closeFn)
	if err != nil {
		return nil, fmt.Errorf("error init rate limiter: %w", err)
	}
	repos, err := deps.NewRepos(config.Storage, dbPool)
	if err != nil {
		return nil, fmt.Errorf("error init data layer %w", err)
	}
	dependencies := &deps.Deps{
		Repos:       repos,
		Logger:      logs,
		RateLimiter: rateLimiter,
		Clock:       clock.New(),
	}

	services := deps.NewServices(dependencies, config)

	return &BruteFP{
		config:   config,
		closer:   closes,
		services: services,
		deps:     dependencies,
		logger:   logs,
	}, nil
}

func (bfp *BruteFP) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	grpcServer, closeFn := grpc.NewHandledServer(bfp.config.API, bfp.services, bfp.deps)
	bfp.closer.Register("GRPC Server", closeFn)

	go func() {
		bfp.logger.Info("GRPC server starting")
		if err := grpcServer.Start(); err != nil {
			bfp.logger.Error("failed to start GRPC server: %w", err)
			cancel()
		}
	}()

	bfp.logger.Info("BruteFP is running...")
	<-ctx.Done()

	return nil
}

func (bfp *BruteFP) Close() {
	// 10 секунд на завершение
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	bfp.closer.Close(ctx, bfp.logger)
	bfp.logger.Info("BruteFP stopped")
}
