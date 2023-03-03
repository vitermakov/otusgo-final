package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"github.com/vitermakov/otusgo-final/internal/app"
	config "github.com/vitermakov/otusgo-final/internal/app/config/brutefp-cli"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/brute_fp/config-cli.json", "Path to configuration file")
}

func main() {
	flag.Parse()
	if flag.Arg(0) == "version" {
		printVersion()
		return
	}
	cfg, err := config.New(configFile)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	appCli, err := app.NewBruteFPCli(ctx, cfg)
	if err != nil {
		log.Printf("can't initialize application: %s\n", err)
		cancel()
		return
	}
	app.Execute(ctx, appCli)
}
