package service

import (
	"context"
	"errors"
	"net"

	"github.com/vitermakov/otusgo-final/internal/model"
	"github.com/vitermakov/otusgo-final/internal/repository"
	"github.com/vitermakov/otusgo-final/pkg/utils/errx"
)

type IPRuleSrv struct {
	repo repository.IPRule
}

func (irs IPRuleSrv) validateAdd(ctx context.Context, input model.IPRuleInput) error {
	if err := input.Validate(); err != nil {
		return err
	}
	rules, err := irs.repo.GetList(ctx, model.IPRuleSearch{
		IPNet:      &input.IPNet,
		IPNetExact: true,
	})
	if err != nil {
		return errx.FatalNew(err)
	}
	if len(rules) > 0 {
		return errx.LogicNew(model.ErrIPRuleNetDuplicate, model.ErrIPRuleNetDuplicateCode)
	}
	return nil
}

func (irs IPRuleSrv) Add(ctx context.Context, input model.IPRuleInput) (*model.IPRule, error) {
	if err := irs.validateAdd(ctx, input); err != nil {
		errs := errx.NamedErrors{}
		if errors.As(err, &errs) {
			return nil, errx.InvalidNew("неверные параметры", errs)
		}
		return nil, err
	}
	ipRule, err := irs.repo.Add(ctx, input)
	if err != nil {
		return nil, errx.FatalNew(err)
	}
	return ipRule, nil
}

func (irs IPRuleSrv) Delete(ctx context.Context, rule model.IPRule) error {
	if err := irs.repo.Delete(ctx, model.IPRuleInput{Type: rule.Type, IPNet: rule.IPNet}); err != nil {
		// неустранимая пользователем ошибка.
		return errx.FatalNew(err)
	}
	return nil
}

// GetRuleTypeForIP по условию, если ip подходит к white списку, то разрешаем без проверки ограничения black.
func (irs IPRuleSrv) GetRuleTypeForIP(ctx context.Context, ip net.IP) (model.RuleType, error) {
	rules, err := irs.repo.GetList(ctx, model.IPRuleSearch{
		IPNet: &net.IPNet{
			IP:   ip,
			Mask: []byte{255, 255, 255, 255},
		},
		IPNetExact: false,
	})
	if err != nil {
		return model.RuleTypeNone, errx.FatalNew(err)
	}
	var deny bool
	for _, rule := range rules {
		if rule.Type == model.RuleTypeAllow {
			return model.RuleTypeAllow, nil
		}
		if rule.Type == model.RuleTypeDeny {
			deny = true
		}
	}
	if deny {
		return model.RuleTypeDeny, nil
	}
	return model.RuleTypeNone, nil
}

func (irs IPRuleSrv) GetByIPNet(ctx context.Context, typ model.RuleType, ipNet net.IPNet) (*model.IPRule, error) {
	event, err := irs.getOne(ctx, model.IPRuleSearch{IPNet: &ipNet, IPNetExact: true})
	if err == nil {
		return event, nil
	}
	// если ошибка - NotFound, добавим параметр ip_net=ipNet.
	nfErr := errx.NotFound{}
	if errors.As(err, &nfErr) {
		nfErr.Params = map[string]interface{}{"ip_net": ipNet.String()}
		return nil, nfErr
	}
	return nil, err
}

func (irs IPRuleSrv) getOne(ctx context.Context, search model.IPRuleSearch) (*model.IPRule, error) {
	rules, err := irs.repo.GetList(ctx, search)
	if err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return nil, errx.NotFoundNew(model.ErrRuleNotFound, nil)
	}
	return &rules[0], nil
}

func NewIPRuleSrv(repo repository.IPRule) IPRule {
	return &IPRuleSrv{
		repo: repo,
	}
}
