package errx

import (
	"fmt"
	"net/http"
)

const (
	TypeNone     = iota
	TypeLogic    // логическая ошибка
	TypePerms    // ошибка прав доступа
	TypeNotFound // объект не найден
	TypeInvalid  // ошибка валидации
	TypeFatal    // критическая внешняя ошибка
)

type Base struct {
	err  error
	kind byte
}

func (err Base) Error() string {
	return err.err.Error()
}

func (err Base) Kind() byte {
	return err.kind
}

// Logic ошибка бизнес-логики (устранимая) приложения.
type Logic struct {
	Base
	code int // код внутренней классификации ошибок.
}

func (err Logic) Code() int {
	return err.code
}

func LogicNew(err error, code int) Logic {
	if code <= 0 {
		code = http.StatusBadRequest
	}
	return Logic{Base{err, TypeLogic}, code}
}

func PermsNew(err error) Base {
	return Base{err, TypePerms}
}

// NotFound объект не найден по какому-то набору для поиска.
type NotFound struct {
	Base
	Params interface{} // параметры поиска
}

func NotFoundNew(err error, params interface{}) NotFound {
	return NotFound{Base{err, TypeNotFound}, params}
}

func FatalNew(err error) Base {
	return Base{err, TypeFatal}
}

type Invalid struct {
	title  string
	errors NamedErrors
}

func (err Invalid) Error() string {
	if !err.Fails() {
		return ""
	}
	return fmt.Sprintf("%s: %s", err.title, err.errors.Error())
}

func (err Invalid) Errors() []NamedError {
	return err.errors
}

func (err Invalid) Fails() bool {
	return !err.errors.Empty()
}

func (err Invalid) Kind() byte {
	return TypeInvalid
}

func InvalidNew(title string, errors NamedErrors) Invalid {
	return Invalid{title: title, errors: errors}
}
