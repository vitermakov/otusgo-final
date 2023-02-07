package model

import "errors"

var (
	ErrDeniedInternal        = errors.New("сервис не работает")
	ErrDeniedByRule          = errors.New("IP находится в black-list")
	ErrDeniedByLoginLimit    = errors.New("достигнут предел запросов для логина")
	ErrDeniedByPasswordLimit = errors.New("достигнут предел запросов для пароля")
	ErrDeniedByIPLimit       = errors.New("достигнут предел для IP")
	ErrWrongResetName        = errors.New(`неизвестный параметр сброса бакета. Ожидается 'login' или 'ip'`)
)

const (
	ErrDeniedInternalCode        = 3001
	ErrDeniedByRuleCode          = 3002
	ErrDeniedByLoginLimitCode    = 3003
	ErrDeniedByPasswordLimitCode = 3004
	ErrDeniedByIPLimitCode       = 3005
	ErrWrongResetNameCode        = 3006
)
