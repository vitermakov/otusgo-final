package app

import (
	"context"
	stdlog "log"

	_ "github.com/jackc/pgx/v4/stdlib" // pgx driver for database/sql
)

type App interface {
	Initialize(ctx context.Context) error
	Run(ctx context.Context) error
	Close()
}

// Execute шаблонная функция выполнения приложения.
func Execute(ctx context.Context, app App) {
	// пропишем defer на закрытие приложения до инициализации.
	defer app.Close()

	if err := app.Initialize(ctx); err != nil {
		stdlog.Fatalf("не удалось инициализировать приложение: %s", err)
	}
	if err := app.Run(ctx); err != nil {
		stdlog.Fatalf("не удалось запустить приложение: %s", err)
	}
}
