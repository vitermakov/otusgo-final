package closer

import (
	"context"
	"errors"
	"sync"

	"github.com/vitermakov/otusgo-final/pkg/logger"
)

var ErrShutdownCanceled = errors.New("shutdown cancelled")

// CloseFunc функция для завершения сервиса.
type CloseFunc func(ctx context.Context) error

type Closer struct {
	mu        sync.Mutex
	closeFunc map[string]CloseFunc
}

func NewCloser() *Closer {
	return &Closer{closeFunc: make(map[string]CloseFunc)}
}

func (c *Closer) Register(name string, closeFunc CloseFunc) {
	if closeFunc == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closeFunc[name] = closeFunc
}

func (c *Closer) Close(ctx context.Context, logger logger.Logger) {
	complete := make(chan struct{}, 1)
	c.mu.Lock()
	defer c.mu.Unlock()

	go func() {
		defer close(complete)
		for name, closer := range c.closeFunc {
			logger.Info("%s: closing", name)
			if err := closer(ctx); err != nil {
				logger.Error("error closing %s: %s", name, err.Error())
			}
			logger.Info("%s: closed successfully", name)
		}
	}()

	select {
	case <-complete:
		break
	case <-ctx.Done():
		close(complete)
		logger.Error("closer exited by timeout")
	}
}
