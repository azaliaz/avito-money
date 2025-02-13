package storage

import (
	"fmt"
	"log/slog"
	"net"
	"time"
)

const (
	defaultHost = "localhost"
	defaultPort = "5432"
)

type Config struct {
	Host             string        `env:"HOST" yaml:"host"`
	DbName           string        `env:"NAME"     envDefault:"postgres"  yaml:"name"`
	User             string        `env:"USER"     envDefault:"user"      yaml:"user"`
	Password         string        `env:"PASSWORD" yaml:"password"`
	MaxOpenConns     int32         `env:"MAX_OPEN_CONNS" envDefault:"10" yaml:"max-open-conns"`
	ConnIdleLifetime time.Duration `env:"CONN_IDLE_LIFETIME" envDefault:"10m" yaml:"conn-idle-lifetime"`
	ConnMaxLifetime  time.Duration `env:"CONN_MAX_LIFETIME" envDefault:"1h" yaml:"conn-max-lifetime"`
}

func (config Config) dsnPostgres(log *slog.Logger) string {
	host, port, err := net.SplitHostPort(config.Host)
	if err != nil {
		log.Error("parse db connect settings", slog.String("err", err.Error()))
		host = defaultHost
		port = defaultPort
	}

	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, config.User, config.Password, config.DbName, port)
}

func (config Config) UrlPostgres() string {
	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		config.User, config.Password, config.Host, config.DbName)
}
