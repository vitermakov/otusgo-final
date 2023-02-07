package pgconn

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"net"
	"net/url"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib" //nolint:blank-imports // не понимаю о чем тут речь
	common "github.com/vitermakov/otusgo-final/internal/app/config"
	"github.com/vitermakov/otusgo-final/pkg/logger"
	"github.com/vitermakov/otusgo-final/pkg/utils/closer"
)

const (
	defConnAttemptsWait = 2
	defConnMaxLifetime  = 20
	defConnMaxIdleTime  = 1
	defMaxOpenCons      = 10
	defMaxIdleCons      = 40
)

func NewPgConn(appName string, config common.SQLConn, log logger.Logger) (*sql.DB, closer.CloseFunc) {
	dbHost := net.JoinHostPort(config.Host, strconv.Itoa(config.Port))
	dsnURL := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(config.User, config.Password),
		Host:     dbHost,
		Path:     "/" + config.DBName,
		RawQuery: "application_name=" + appName,
	}

	var dbPool *sql.DB
	attempts := config.ConnAttemptsCount
	if attempts <= 0 {
		attempts = 1
	}
	if config.ConnAttemptsWait <= 0 {
		config.ConnAttemptsWait = defConnAttemptsWait
	}

	for i := 1; i <= attempts; i++ {
		log.Info("connecting DB: %s, attempt: %d/%d", dbHost, i, attempts)
		pool, err := sql.Open("pgx", dsnURL.String())
		if err == nil {
			dbPool = pool
			break
		}
		// продолжаем попытки только при ErrBadConn
		if !errors.Is(err, driver.ErrBadConn) {
			log.Error("can't connecting DB to host %s: %s", dbHost, err.Error())
			return nil, nil
		}
		if i == attempts {
			log.Error("stop attempting connecting DB to host %s", dbHost, err.Error())
			return nil, nil
		}

		time.Sleep(time.Duration(config.ConnAttemptsWait) * time.Second)
	}

	if config.ConnMaxLifetime <= 0 {
		config.ConnMaxLifetime = defConnMaxLifetime
	}
	dbPool.SetConnMaxLifetime(time.Duration(config.ConnMaxLifetime) * time.Second)

	if config.ConnMaxIdleTime <= 0 {
		config.ConnMaxIdleTime = defConnMaxIdleTime
	}
	dbPool.SetConnMaxIdleTime(time.Duration(config.ConnMaxIdleTime) * time.Second)

	if config.MaxOpenCons <= 0 {
		config.MaxOpenCons = defMaxOpenCons
	}
	dbPool.SetMaxOpenConns(config.MaxOpenCons)

	if config.MaxIdleCons <= 0 {
		config.MaxIdleCons = defMaxIdleCons
	}
	dbPool.SetMaxIdleConns(config.MaxIdleCons)

	log.Info("DB connected to %s", dbHost)

	return dbPool, func(_ context.Context) error {
		return dbPool.Close()
	}
}
