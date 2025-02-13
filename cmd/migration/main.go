package main

import (
	"flag"
	"github.com/azaliaz/avito-money/internal/storage"
	"github.com/azaliaz/avito-money/migrations"
	"github.com/azaliaz/avito-money/pkg/config"
	"log/slog"
	"os"
)

var version string

type Config struct {
	DBConfig storage.Config `envPrefix:"DB_" yaml:"db-config"`
	Path     string         `env:"PATH" yaml:"path"`
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
		logger.Error("config parse error:", err)
		os.Exit(1)
	}

	if err := migrations.PostgresMigrate(cfg.DBConfig.UrlPostgres()); err != nil {
		logger.Error("migration error", err)
		os.Exit(1)
	}
	logger.Info("migration completed", slog.String("path", cfg.Path))

}
