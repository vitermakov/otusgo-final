package repository

import (
	"context"

	"github.com/vitermakov/otusgo-final/internal/model"
)

// IPRule управление хранилищем white/black списков.
type IPRule interface {
	Add(context.Context, model.IPRuleInput) (*model.IPRule, error)
	Delete(context.Context, model.IPRuleInput) error
	GetList(context.Context, model.IPRuleSearch) ([]model.IPRule, error)
}
