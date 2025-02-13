package main

import (
	"context"
	"flag"
	"github.com/azaliaz/avito-shop/internal/application"
	"github.com/azaliaz/avito-shop/internal/facade/rest"
	"github.com/azaliaz/avito-shop/internal/storage"
	"github.com/azaliaz/avito-shop/pkg/config"
	"github.com/azaliaz/avito-shop/pkg/service"
	"log/slog"
	"os"
)

var version string

type Config struct {
	App     application.Config `envPrefix:"APP_" yaml:"app"`
	Storage storage.Config     `envPrefix:"STORAGE_" yaml:"storage"`
	Rest    rest.Config        `envPrefix:"REST_" yaml:"rest"`
}

func main() {
	/* Configuring logger */
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
	/* Configuring flags */
	configFile := flag.String("config-file", "none", "config file")
	flag.Parse()

	/* Parsing config */
	cfg := Config{}
	err := config.ReadConfig(*configFile, &cfg)
	if err != nil {
		logger.Error("config parse error:", "err_msg", err)
		return
	}

	db := storage.NewDB(&cfg.Storage, logger)
	repo := storage.NewService(db, logger)
	app := application.NewService(logger, &cfg.App, repo)
	api := rest.NewAPI(logger, &cfg.Rest, app)

	mgr := service.NewManager(logger)
	mgr.AddService(db, app, api)

	ctx := context.Background()
	if err := mgr.Run(ctx); err != nil {
		logger.Error("can't start services:", err)
	}
}
