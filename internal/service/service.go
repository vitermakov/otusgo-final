package service

import (
	"context"
	"net"
	"time"

	"github.com/vitermakov/otusgo-final/internal/model"
)

// IPRule управление white/black списками.
type IPRule interface {
	Add(context.Context, model.IPRuleInput) (*model.IPRule, error)
	Delete(context.Context, model.IPRule) error
	GetByIPNet(context.Context, model.RuleType, net.IPNet) (*model.IPRule, error)
	GetRuleTypeForIP(context.Context, net.IP) (model.RuleType, error)
}

// PermitChecker проверка разрешения на совершение действия, основываясь на политике лимитов.
type PermitChecker interface {
	Check(context.Context, model.PermitQuery) (model.PermitResult, error)
	Reset(context.Context, model.ResetQuery) (bool, error)
	SetBaseDuration(time.Duration)
}
