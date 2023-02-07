package model

import (
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/vitermakov/otusgo-final/pkg/utils/errx"
)

type RuleType int

const (
	RuleTypeNone RuleType = iota
	RuleTypeAllow
	RuleTypeDeny
	RuleTypeError
)

func (lit RuleType) Valid() bool {
	return lit > RuleTypeNone && lit < RuleTypeError
}

func (lit RuleType) String() string {
	switch lit { //nolint:exhaustive // и не должно быть
	case RuleTypeAllow:
		return "allow"
	case RuleTypeDeny:
		return "deny"
	}
	return ""
}

func ParseRuleType(ruleType string) (RuleType, error) {
	switch ruleType {
	case "allow":
		return RuleTypeAllow, nil
	case "deny":
		return RuleTypeDeny, nil
	}
	return RuleTypeNone, ErrRuleTypeUnk
}

// IPRule сущность элемента white/black листов.
// На данный момент мы не предусматриваем порядок применения правил.
type IPRule struct {
	ID        uuid.UUID
	Type      RuleType
	IPNet     net.IPNet
	UpdatedAt time.Time
}

// IPRuleInput структура для добавления или удаления подсети из white/black листов.
type IPRuleInput struct {
	Type  RuleType
	IPNet net.IPNet
}

func (rc IPRuleInput) Validate() error {
	var errs errx.NamedErrors
	if !rc.Type.Valid() {
		errs.Add(errx.NamedError{
			Field: "Type",
			Err:   ErrRuleTypeUnk,
		})
	}
	if errs.Empty() {
		return nil
	}
	return errs
}

// IPRuleSearch структура для поиска правил.
type IPRuleSearch struct {
	// UUID правила
	ID *uuid.UUID
	// Type тип правила
	Type *RuleType
	// IP/подсеть
	IPNet *net.IPNet
	// IPNetExact точное соответствие (false - значит IPNet IP/подсеть входит в указанную в правиле)
	IPNetExact bool
}
