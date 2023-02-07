package model

import "errors"

var (
	ErrRuleTypeUnk  = errors.New("неизвестный тип правила (allow/deny)")
	ErrRuleNotFound = errors.New("правило не найдено")
)

const (
	ErrIPRuleNetDuplicateCode = 2001
)

var ErrIPRuleNetDuplicate = errors.New("указанная сеть уже добавлена")
