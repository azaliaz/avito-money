package rest

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/azaliaz/avito-money/internal/application"
	"log/slog"
	"time"
)

type Service struct {
	log    *slog.Logger
	config *Config
	fiber  *fiber.App
	app    application.ShopService
}

func NewAPI(
	logEntry *slog.Logger,
	config *Config,
	app application.ShopService,
) *Service {
	return &Service{
		log:    logEntry,
		config: config,
		app:    app,
	}
}

func (api *Service) Init() error {
	api.fiber = fiber.New(fiber.Config{
		ReadTimeout:           time.Duration(api.config.FiberReadTimeout) * time.Second,
		WriteTimeout:          time.Duration(api.config.FiberWriteTimeout) * time.Second,
		IdleTimeout:           time.Duration(api.config.FiberIdleTimeout) * time.Second,
		BodyLimit:             int(api.config.FiberBodyLimit),
		ReadBufferSize:        int(api.config.FiberReadBufferSize),
		StrictRouting:         api.config.FiberStrictRouting,
		CaseSensitive:         api.config.FiberCaseSensitive,
		DisableStartupMessage: api.config.FiberDisableStartupMessage,
		DisableKeepalive:      api.config.FiberDisableKeepalive,
	})

	api.fiber.Add("POST", "/api/auth", api.Auth)
	api.fiber.Add("GET", "/api/buy/:item", api.BuyItem)
	api.fiber.Add("GET", "/api/info", api.Info)
	api.fiber.Add("POST", "/api/sendCoin", api.SendCoin)

	fmt.Println(api.config.Port)
	addr := fmt.Sprintf(":%d", api.config.Port)
	err := api.fiber.Listen(addr)
	if err != nil {
		return err
	}

	return nil
}

func (api *Service) Run(ctx context.Context) {
	addr := fmt.Sprintf(":%d", api.config.Port)
	api.log.Info("start rest server", "addr", addr)
	if err := api.fiber.Listen(addr); err != nil {
		api.log.Error("start rest server", "addr", addr, "portal", err)
		return
	}
}

func (api *Service) Stop() {
}
