package brutefp

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/benbjohnson/clock"
	common "github.com/vitermakov/otusgo-final/internal/app/config"
	"github.com/vitermakov/otusgo-final/internal/app/config/brutefp"
	"github.com/vitermakov/otusgo-final/internal/ratelimit"
	"github.com/vitermakov/otusgo-final/internal/repository"
	"github.com/vitermakov/otusgo-final/internal/repository/memory"
	"github.com/vitermakov/otusgo-final/internal/repository/pgsql"
	"github.com/vitermakov/otusgo-final/internal/service"
	"github.com/vitermakov/otusgo-final/pkg/logger"
)

const (
	StoreTypeInMemory = "memory"
	StoreTypeInPgsql  = "pgsql"
)

// Repos регистр репозиториев.
type Repos struct {
	IPRule repository.IPRule
}

func NewRepos(store common.Storage, dbPool *sql.DB) (*Repos, error) {
	var (
		repos = &Repos{}
		err   error
	)
	switch store.Type {
	case StoreTypeInMemory:
		repos = &Repos{
			IPRule: memory.NewIPRuleRepo(),
		}
	case StoreTypeInPgsql:
		repos = &Repos{
			IPRule: pgsql.NewIPRuleRepo(dbPool),
		}
	default:
		err = fmt.Errorf("unknown storage type '%s", store.Type)
	}
	return repos, err
}

// Deps зависимости.
type Deps struct {
	Repos       *Repos
	Logger      logger.Logger
	RateLimiter ratelimit.RateLimiter
	Clock       clock.Clock
}

// Services регистр сервисов.
type Services struct {
	IPRule        service.IPRule
	PermitChecker service.PermitChecker
}

func NewServices(deps *Deps, cfg brutefp.Config) *Services {
	repos := deps.Repos
	ipRule := service.NewIPRuleSrv(repos.IPRule)
	return &Services{
		IPRule:        ipRule,
		PermitChecker: service.NewPermitCheckerSrv(ipRule, deps.RateLimiter, time.Minute, cfg.Limits),
	}
}
