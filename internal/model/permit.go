package model

import (
	"net"
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
}

// ResetQuery запрос на сброс блокировки.
type ResetQuery struct {
	Name  string
	Value string
}

func (rq ResetQuery) Valid() bool {
	return rq.Name == "login" || rq.Name == "ip"
}
