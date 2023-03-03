package model

import "errors"

var (
	ErrDeniedInternal        = errors.New("сервис не работает")
	ErrDeniedByRule          = errors.New("the IP is in the black-list")
	ErrDeniedByLoginLimit    = errors.New("the limit of requests has been reached (login)")
	ErrDeniedByPasswordLimit = errors.New("the limit of requests has been reached (password)")
	ErrDeniedByIPLimit       = errors.New("the limit of requests has been reached (ip)")
	ErrWrongResetName        = errors.New(`unknown bucket reset parameter. 'login' или 'ip' expected`)
)

const (
	ErrDeniedInternalCode        = 3001
	ErrDeniedByRuleCode          = 3002
	ErrDeniedByLoginLimitCode    = 3003
	ErrDeniedByPasswordLimitCode = 3004
	ErrDeniedByIPLimitCode       = 3005
	ErrWrongResetNameCode        = 3006
)
