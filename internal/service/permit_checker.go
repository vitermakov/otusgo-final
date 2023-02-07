package service

import (
	"context"
	"fmt"
	"time"

	config "github.com/vitermakov/otusgo-final/internal/app/config/brutefp"
	"github.com/vitermakov/otusgo-final/internal/model"
	"github.com/vitermakov/otusgo-final/internal/ratelimit"
	"github.com/vitermakov/otusgo-final/pkg/utils/errx"
)

type PermitCheckerSrv struct {
	ipRule       IPRule
	rateLimiter  ratelimit.Limiter
	baseDuration time.Duration
	limits       config.Limits
}

func (p PermitCheckerSrv) Check(ctx context.Context, query model.PermitQuery) (model.PermitResult, error) {
	// сначала проверяем, если ли IP в white/black списках
	ruleType, err := p.ipRule.GetRuleTypeForIP(ctx, query.IP)
	if err != nil {
		return model.PermitResult{Err: model.ErrDeniedInternal}, err
	}
	if ruleType.Valid() {
		switch ruleType { //nolint:exhaustive // и не должно быть
		case model.RuleTypeAllow:
			return model.PermitResult{Success: true}, nil
		case model.RuleTypeDeny:
			return model.PermitResult{Err: fmt.Errorf("%w: %s", model.ErrDeniedByRule, query.IP)}, nil
		}
	}

	// проверки на лимиты. Для каждого типа лимита свой ключ
	checks := []struct {
		name  string
		limit int
		err   error
	}{
		{
			name:  fmt.Sprintf("login_%s", query.Login),
			limit: p.limits.LoginPerMin,
			err:   fmt.Errorf("%w: %s", model.ErrDeniedByLoginLimit, query.Login),
		}, {
			name:  fmt.Sprintf("password_%s", query.Password),
			limit: p.limits.PasswordPerMin,
			err:   fmt.Errorf("%w: %s", model.ErrDeniedByPasswordLimit, query.Password),
		}, {
			name:  fmt.Sprintf("ip_%s", query.IP),
			limit: p.limits.IPPerMin,
			err:   fmt.Errorf("%w: %s", model.ErrDeniedByIPLimit, query.IP),
		},
	}

	var exceed bool
	for _, check := range checks {
		exceed, err = p.rateLimiter.ExceedLimit(check.name, ratelimit.Limits{
			Period: p.baseDuration,
			Limit:  int64(check.limit),
		})
		if err != nil {
			return model.PermitResult{Err: model.ErrDeniedInternal}, errx.FatalNew(err)
		}
		if exceed {
			return model.PermitResult{Err: check.err}, nil
		}
	}

	return model.PermitResult{Success: true}, nil
}

func (p PermitCheckerSrv) Reset(ctx context.Context, query model.ResetQuery) (bool, error) {
	if !query.Valid() {
		return false, errx.LogicNew(model.ErrWrongResetName, model.ErrWrongResetNameCode)
	}
	found, err := p.rateLimiter.ResetBucket(fmt.Sprintf("%s_%s", query.Name, query.Value))
	if err != nil {
		return false, errx.FatalNew(err)
	}
	return found, nil
}

func (p *PermitCheckerSrv) SetBaseDuration(bd time.Duration) {
	p.baseDuration = bd
}

func NewPermitCheckerSrv(
	ipRule IPRule, limiter ratelimit.RateLimiter, bd time.Duration, cfg config.Limits,
) PermitChecker {
	return &PermitCheckerSrv{ipRule: ipRule, rateLimiter: limiter, baseDuration: bd, limits: cfg}
}
