package model

import (
	"net"
)

const (
	LimitParamNameLogin    = "login"
	LimitParamNamePassword = "password"
	LimitParamNameIP       = "ip"
)

// PermitQuery запрос проверки параметров авторизации.
type PermitQuery struct {
	Login    string
	Password string
	IP       net.IP
}

// PermitResult результат проверки на bruteforce.
type PermitResult struct {
	Success bool
	Err     error
	ErrCode int
}

// LimitBucket запрос на сброс блокировки.
type LimitBucket struct {
	Param string
	Value string
}

func (lb LimitBucket) Valid() bool {
	return lb.ValidForReset() || lb.Param == LimitParamNamePassword
}

func (lb LimitBucket) ValidForReset() bool {
	return lb.Param == LimitParamNameLogin || lb.Param == LimitParamNameIP
}

func (lb LimitBucket) Name() string {
	return lb.Param + "_" + lb.Value
}

func (lb LimitBucket) String() string {
	return lb.Param + ": " + lb.Value
}
