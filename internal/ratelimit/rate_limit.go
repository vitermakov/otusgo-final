package ratelimit

import (
	"context"
	"errors"
	"time"

	config "github.com/vitermakov/otusgo-final/internal/app/config/brutefp"
	"github.com/vitermakov/otusgo-final/pkg/utils/closer"
)

var (
	ErrInvalidLimitOpts = errors.New("invalid rate limits")
	ErrMethodUnknown    = errors.New("unknown rate limit method")
)

// Limits настройки ограничения скорости.
type Limits struct {
	Period time.Duration
	Limit  int64
}

func (c Limits) Valid() error {
	if c.Period.Nanoseconds() > 0 && c.Limit > 0 {
		return nil
	}
	return ErrInvalidLimitOpts
}

// Limiter интерфейс ограничителя частоты запросов.
type Limiter interface {
	// ExceedLimit метод проверяет превышает ли новое входящее событие ограничение
	// и учитывает его в бакете с кодом bucketCode
	ExceedLimit(bucketCode string, config Limits) (bool, error)
	// ResetBucket сбрасывает бакет
	ResetBucket(bucketCode string) (bool, error)
}

// RateLimiter интерфейс сервиса-ограничителя частоты запросов.
type RateLimiter interface {
	Limiter
	Destroy() error
}

func NewRateLimiter(cfg config.Limits) (RateLimiter, closer.CloseFunc, error) {
	var limiter RateLimiter
	if cfg.Method == "fixed_memory" {
		limiter = NewFixedMemory()
	}
	if limiter != nil {
		return limiter, func(ctx context.Context) error {
			return limiter.Destroy()
		}, nil
	}
	return nil, nil, ErrMethodUnknown
}
