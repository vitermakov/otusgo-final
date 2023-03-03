package model

import "errors"

var (
	ErrRuleTypeUnk  = errors.New("unknown rule type (allow/deny)")
	ErrRuleNotFound = errors.New("rule fot specified network not found")
)

const (
	ErrIPRuleNetDuplicateCode = 2001
)

var ErrIPRuleNetDuplicate = errors.New("specified network already exists")
