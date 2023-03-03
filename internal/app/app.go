package app

import (
	"context"
	stdlog "log"

	_ "github.com/jackc/pgx/v4/stdlib" // pgx driver for database/sql
)

type App interface {
	Run(ctx context.Context) error
	Close()
}

// Execute шаблонная функция выполнения приложения.
func Execute(ctx context.Context, app App) {
	// здесь создадим собственный контекс для того, чтобы иметь возможность отменить переданный
	// в app.Initialize и app.Run контекс внутри функции Execute
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// пропишем defer на закрытие приложения до инициализации.
	defer app.Close()

	if err := app.Run(ctx); err != nil {
		stdlog.Printf("can't run application: %s\n", err)
		cancel()
		return
	}
}
