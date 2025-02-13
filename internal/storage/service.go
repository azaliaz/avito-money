package storage

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"time"
)

type ShopStorage interface {
	Auth(ctx context.Context, request *AuthRequest) (*AuthResponse, error)
	GetInventory(ctx context.Context, userId uint64) ([]*ProductStock, error)
	GetBalance(ctx context.Context, userId uint64) (balance int, err error)
	GetCoinHistory(ctx context.Context, userId uint64) (*CoinHistory, error)
	SendCoin(ctx context.Context, request *SendCoinRequest) (*SendCoinResponse, error)
	BuyItem(ctx context.Context, request *BuyItemRequest) (*BuyItemResponse, error)
}

type AuthRequest struct {
	UserName string
	PassHash string
}

type AuthResponse struct {
	UserId   uint64
	UserName string
	PassHash string
}

type CoinHistory struct {
	Received []*Transaction
	Sent     []*Transaction
}

type Transaction struct {
	Amount    int
	FromUser  uint64
	ToUser    uint64
	CreatedAt time.Time
}

type ProductStock struct {
	Type     string
	Quantity int
}

type SendCoinRequest struct {
	UserId uint64
	Amount int
	ToUser uint64
}

type SendCoinResponse struct{}

type BuyItemRequest struct {
	UserId uint64
	Item   string
}

type BuyItemResponse struct {
}

func NewService(db *DB, logger *slog.Logger) *Service {
	return &Service{DB: db, logger: logger}
}

type Service struct {
	*DB
	logger *slog.Logger
}

func NewDB(config *Config, logEntry *slog.Logger) *DB {
	return &DB{
		config: config,
		log:    logEntry,
	}
}

type DB struct {
	config *Config
	log    *slog.Logger
	pool   *pgxpool.Pool
	cancel func()
}

func (r *DB) Init() error {
	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel

	poolCfg, err := pgxpool.ParseConfig(r.config.dsnPostgres(r.log))
	if err != nil {
		return fmt.Errorf("error on parsing rw storage config: %w", err)
	}

	poolCfg.MaxConns = r.config.MaxOpenConns
	poolCfg.MaxConnIdleTime = r.config.ConnIdleLifetime
	poolCfg.MaxConnLifetime = r.config.ConnMaxLifetime

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return fmt.Errorf("error on creating rw storage connection pool: %w", err)
	}

	r.pool = pool

	r.log.Info("connected to postgres")
	return nil
}

func (r *DB) Run(_ context.Context) {
}

func (r *DB) Stop() {
	r.log.Info("stopping storage service")
	if r.cancel != nil {
		r.cancel()
	}
	r.pool.Close()
	r.log.Info("storage service has been stopped")
}

func (r *DB) Pool() *pgxpool.Pool {
	return r.pool
}
