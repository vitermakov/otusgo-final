package servers

import (
	"context"
)

// AuthUser авторизованный пользователь.
type AuthUser struct {
	ID    string
	Login string
	Name  string
}

// AuthService интерфейс микросервиса авторизации.
type AuthService interface {
	Authorize(context.Context, string) (*AuthUser, error)
}
