package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	_ "time/tzdata"

	"github.com/Jinvic/hookhub/internal/app"
)

// Set via -ldflags "-X main.version=..." at release build time.
var version = "dev"

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	bootstrap := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := app.LoadConfig(*configPath)
	if err != nil {
		bootstrap.Error("load config", "err", err)
		os.Exit(1)
	}

	log, err := app.NewLogger(cfg.Server.LogLevel)
	if err != nil {
		bootstrap.Error("init logger", "err", err)
		os.Exit(1)
	}

	application, err := app.New(cfg, log)
	if err != nil {
		log.Error("init app", "err", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := application.Run(ctx); err != nil {
		log.Error("run", "err", err)
		os.Exit(1)
	}
}
