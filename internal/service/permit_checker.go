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
	checks       []struct {
		name    string
		limit   int
		err     error
		errCode int
	}
}

func (p PermitCheckerSrv) Check(ctx context.Context, query model.PermitQuery) (model.PermitResult, error) {
	// сначала проверяем, если ли IP в white/black списках
	ruleType, err := p.ipRule.GetRuleTypeForIP(ctx, query.IP)
	if err != nil {
		return model.PermitResult{Err: model.ErrDeniedInternal, ErrCode: model.ErrDeniedInternalCode}, err
	}
	if ruleType.Valid() {
		switch ruleType { //nolint:exhaustive // и не должно быть
		case model.RuleTypeAllow:
			return model.PermitResult{Success: true}, nil
		case model.RuleTypeDeny:
			return model.PermitResult{
				Err:     fmt.Errorf("%w: %s", model.ErrDeniedByRule, query.IP),
				ErrCode: model.ErrDeniedByRuleCode,
			}, nil
		}
	}

	var exceed bool
	for _, check := range p.checks {
		var checkValue string
		switch check.name {
		case model.LimitParamNameLogin:
			checkValue = query.Login
		case model.LimitParamNamePassword:
			checkValue = query.Password
		case model.LimitParamNameIP:
			checkValue = query.IP.String()
		}
		exceed, err = p.rateLimiter.ExceedLimit(
			model.LimitBucket{Param: check.name, Value: checkValue}.Name(),
			ratelimit.Limits{
				Period: p.baseDuration,
				Limit:  int64(check.limit),
			},
		)
		if err != nil {
			return model.PermitResult{Err: model.ErrDeniedInternal}, errx.FatalNew(err)
		}
		if exceed {
			return model.PermitResult{
				Err: fmt.Errorf("%w: %s", check.err, checkValue),
			}, nil
		}
	}

	return model.PermitResult{Success: true}, nil
}

func (p PermitCheckerSrv) Reset(_ context.Context, bucket model.LimitBucket) (bool, error) {
	if !bucket.ValidForReset() {
		return false, errx.LogicNew(model.ErrWrongResetName, model.ErrWrongResetNameCode)
	}
	found, err := p.rateLimiter.ResetBucket(bucket.Name())
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
	// проверки на лимиты. Для каждого типа лимита свой ключ
	checks := []struct {
		name    string
		limit   int
		err     error
		errCode int
	}{
		{
			name:    "login",
			limit:   cfg.LoginPerMin,
			err:     model.ErrDeniedByLoginLimit,
			errCode: model.ErrDeniedByLoginLimitCode,
		}, {
			name:    "password",
			limit:   cfg.PasswordPerMin,
			err:     model.ErrDeniedByPasswordLimit,
			errCode: model.ErrDeniedByPasswordLimitCode,
		}, {
			name:    "ip",
			limit:   cfg.IPPerMin,
			err:     model.ErrDeniedByIPLimit,
			errCode: model.ErrDeniedByIPLimitCode,
		},
	}
	return &PermitCheckerSrv{ipRule: ipRule, rateLimiter: limiter, baseDuration: bd, limits: cfg, checks: checks}
}
