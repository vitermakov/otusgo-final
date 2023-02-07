package memory

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/vitermakov/otusgo-final/internal/model"
	"github.com/vitermakov/otusgo-final/internal/repository"
)

type IPRuleRepo struct {
	mu    sync.RWMutex
	rules []model.IPRule
}

func (ir *IPRuleRepo) Add(_ context.Context, input model.IPRuleInput) (*model.IPRule, error) {
	rule := model.IPRule{
		ID:        uuid.New(),
		Type:      input.Type,
		IPNet:     input.IPNet,
		UpdatedAt: time.Now(),
	}

	ir.mu.Lock()
	ir.rules = append(ir.rules, rule)
	ir.mu.Unlock()

	return &rule, nil
}

func (ir *IPRuleRepo) Delete(ctx context.Context, input model.IPRuleInput) error {
	ir.mu.Lock()
	defer ir.mu.Unlock()
	result := make([]model.IPRule, 0)
	for _, rule := range ir.rules {
		if !ir.matchSearch(rule, model.IPRuleSearch{
			Type:       &input.Type,
			IPNet:      &input.IPNet,
			IPNetExact: true,
		}) {
			result = append(result, rule)
		}
	}
	ir.rules = result
	return nil
}

func (ir *IPRuleRepo) GetList(_ context.Context, search model.IPRuleSearch) ([]model.IPRule, error) {
	var filtered []model.IPRule
	ir.mu.RLock()
	for _, rule := range ir.rules {
		if ir.matchSearch(rule, search) {
			filtered = append(filtered, rule)
		}
	}
	ir.mu.RUnlock()

	return filtered, nil
}

func (ir *IPRuleRepo) matchSearch(rule model.IPRule, search model.IPRuleSearch) bool {
	if search.ID != nil {
		if strings.Compare(rule.ID.String(), search.ID.String()) != 0 {
			return false
		}
	}
	if search.Type != nil {
		if rule.Type != *search.Type {
			return false
		}
	}
	if search.IPNet != nil {
		// если кто-то спросит что это за жесть - линтер считает это более правильным чем вложенные if
		switch search.IPNetExact {
		case true:
			if strings.Compare(rule.IPNet.String(), search.IPNet.String()) != 0 {
				return false
			}
		case false:
			if !rule.IPNet.Contains(search.IPNet.IP) {
				return false
			}
		}
	}
	return true
}

func NewIPRuleRepo() repository.IPRule {
	return &IPRuleRepo{}
}
