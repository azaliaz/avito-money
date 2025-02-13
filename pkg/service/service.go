package service

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"reflect"
)

type (
	Service interface {
		Init() error
		Run(ctx context.Context)
		Stop()
	}
	Services interface {
		AddService(service ...Service)
		Run(ctx context.Context) error
	}
	Manager struct {
		services []Service
		log      *slog.Logger
	}
)

func NewManager(log *slog.Logger) Services {
	return &Manager{log: log}
}

func (s *Manager) AddService(service ...Service) {
	s.services = append(s.services, service...)
}

func (s *Manager) Run(ctx context.Context) (err error) {
	s.log.Info("going to start services")

	// Сможет остановить сервисы если произойдет паника
	defer func() {
		if err != nil {
			s.log.Error("an error occurred", "err", err)
			s.stop()
		}
	}()

	for _, service := range s.services {
		if err = service.Init(); err != nil {
			return fmt.Errorf("failed to run %s: %w", reflect.TypeOf(service), err)
		}
		go service.Run(ctx)
	}

	s.log.Info("the worker has been initialized")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	select {
	case <-c:
		s.log.Info("received an interrupt signal, stopping services")
	case <-ctx.Done():
		s.log.Info("context done, stopping services")
	}

	s.stop()

	return nil
}

func (s *Manager) stop() {
	s.log.Info("going to stop")
	for _, service := range s.services {
		service.Stop()
	}
}
