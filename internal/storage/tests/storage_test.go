package tests

import (
	"context"
	"fmt"
	"github.com/azaliaz/avito-money/internal/storage"
	"github.com/azaliaz/avito-money/migrations"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"log/slog"
	"strconv"
	"testing"
	"time"
)

func (s *RepositoryTestSuite) TestGetBalance() {
	ctx := context.Background()
	userId := uint64(1)
	expectBalance := 800
	prepare := func() {
		conn, err := s.db.Pool().Acquire(ctx)
		require.NoError(s.T(), err)
		defer conn.Release()
		_, err = conn.Exec(ctx, `INSERT INTO users(id, username, password_hash, balance)
					SELECT @user_id, 'abc', 'abc', @balance`,
			pgx.NamedArgs{
				"user_id": userId,
				"balance": expectBalance,
			})
		require.NoError(s.T(), err)
	}
	clear := func() {
		conn, err := s.db.Pool().Acquire(ctx)
		require.NoError(s.T(), err)
		defer conn.Release()
		_, err = conn.Exec(ctx, `DELETE FROM users`)
		require.NoError(s.T(), err)
	}

	prepare()
	fmt.Println("going to get balance")
	balance, err := s.repo.GetBalance(ctx, userId)
	fmt.Println("got balance")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), expectBalance, balance)
	clear()
}


func (s *RepositoryTestSuite) TestAuth() {
    ctx := context.Background()
    username := "test_user"
    passwordHash := "hashed_password"

    prepare := func() {
        conn, err := s.db.Pool().Acquire(ctx)
        require.NoError(s.T(), err)
        defer conn.Release()

       
        _, err = conn.Exec(ctx, `DELETE FROM users WHERE username = $1`, username)
        require.NoError(s.T(), err)
    }
    clear := func() {
        conn, err := s.db.Pool().Acquire(ctx)
        require.NoError(s.T(), err)
        defer conn.Release()
        
        _, err = conn.Exec(ctx, `DELETE FROM users WHERE username = $1`, username)
        require.NoError(s.T(), err)
    }

    authRequest := &storage.AuthRequest{
        UserName:  username,
        PassHash:  passwordHash,
    }

    // Тест аутентификации
    s.T().Run("Authenticate user", func(t *testing.T) {
        prepare()
        resp, err := s.repo.Auth(ctx, authRequest)
        require.NoError(s.T(), err)
        assert.Equal(s.T(), username, resp.UserName)
        assert.NotEmpty(s.T(), resp.PassHash) 
        assert.Greater(s.T(), resp.UserId, uint64(0)) 
    })

    

    // Тест на успешную аутентификацию и создание нового пользователя
    s.T().Run("Authenticate and create new user", func(t *testing.T) {
        authRequest := &storage.AuthRequest{
            UserName:  "new_user", 
            PassHash:  "new_password_hash",
        }

        resp, err := s.repo.Auth(ctx, authRequest)
        require.NoError(s.T(), err)

        
        assert.Equal(s.T(), "new_user", resp.UserName)
        assert.NotEmpty(s.T(), resp.PassHash) 
        assert.Greater(s.T(), resp.UserId, uint64(0)) 
    })
	
	// Тест для аутентификации с неверным паролем
	s.T().Run("Authenticate with incorrect password", func(t *testing.T) {
		prepare()

		
		authRequest := &storage.AuthRequest{
			UserName:  username,
			PassHash:  passwordHash, 
		}

		resp, err := s.repo.Auth(ctx, authRequest)
		require.NoError(s.T(), err)

		
		assert.Equal(s.T(), username, resp.UserName)
		assert.NotEmpty(s.T(), resp.PassHash) 
		assert.Greater(s.T(), resp.UserId, uint64(0)) 

		
		authRequestIncorrect := &storage.AuthRequest{
			UserName:  username,
			PassHash:  "incorrect_password", 
		}

		respIncorrect, err := s.repo.Auth(ctx, authRequestIncorrect)
		
		require.Error(s.T(), err)
		assert.Nil(s.T(), respIncorrect) 
	})

    // Тест аутентификации с пустым паролем
    s.T().Run("Authenticate with empty password", func(t *testing.T) {
        prepare()

        
        authRequest := &storage.AuthRequest{
            UserName:  username,
            PassHash:  "", 
        }

        resp, err := s.repo.Auth(ctx, authRequest)
        require.Error(s.T(), err)
    	assert.Nil(s.T(), resp) 
    })

    // тест для многократной аутентификации с одинаковыми данными
    s.T().Run("Authenticate multiple times with same credentials", func(t *testing.T) {
        prepare()

        
        authRequest := &storage.AuthRequest{
            UserName:  username,
            PassHash:  passwordHash,
        }

        resp1, err := s.repo.Auth(ctx, authRequest)
        require.NoError(s.T(), err)
        assert.Equal(s.T(), username, resp1.UserName)

        resp2, err := s.repo.Auth(ctx, authRequest)
        require.NoError(s.T(), err)
        assert.Equal(s.T(), username, resp2.UserName)
        assert.Equal(s.T(), resp1.UserId, resp2.UserId) 
    })
    
    clear()
}


func TestRepositorySuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}

type RepositoryTestSuite struct {
	container *postgres.PostgresContainer
	suite.Suite

	dbConfig storage.Config

	db   *storage.DB
	repo storage.ShopStorage
}

func (s *RepositoryTestSuite) SetupSuite() {
	ctx := context.Background()
	s.dbConfig = s.setupPostgres(ctx)

	logger := slog.Default()
	db := storage.NewDB(&s.dbConfig, logger)
	if err := db.Init(); err != nil {
		require.NoError(s.T(), err)
	}
	s.db = db
	s.repo = storage.NewService(db, logger)
}

func (s *RepositoryTestSuite) SetupTest() {
    ctx := context.Background()

    
    conn, err := s.db.Pool().Acquire(ctx)
    require.NoError(s.T(), err)
    defer conn.Release()

    
    _, err = conn.Exec(ctx, `DROP SCHEMA public CASCADE; CREATE SCHEMA public;`)
    require.NoError(s.T(), err)

    
    require.NoError(s.T(), migrations.PostgresMigrate(s.dbConfig.UrlPostgres()))
}

func (s *RepositoryTestSuite) TearDownTest() {
	require.NoError(s.T(), migrations.PostgresMigrateDown(s.dbConfig.UrlPostgres()))
}
func (s *RepositoryTestSuite) setupPostgres(ctx context.Context) storage.Config {
	cfg := storage.Config{
		Host:             "",
		DbName:           "test-db",
		User:             "user",
		Password:         "1",
		MaxOpenConns:     10,
		ConnIdleLifetime: 60 * time.Second,
		ConnMaxLifetime:  60 * time.Minute,
	}
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:14-alpine"),
		postgres.WithDatabase(cfg.DbName),
		postgres.WithUsername(cfg.User),
		postgres.WithPassword(cfg.Password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	require.NoError(s.T(), err)
	s.container = pgContainer

	host, err := pgContainer.Host(ctx)
	require.NoError(s.T(), err)
	cfg.Host = host
	ports, err := pgContainer.MappedPort(ctx, "5432")
	require.NoError(s.T(), err)
	cfg.Host += ":" + strconv.Itoa(ports.Int())

	s.dbConfig = cfg
	return cfg
}

func (s *RepositoryTestSuite) TearDownSuite() {
	s.db.Stop()
	s.container.Stop(context.Background(), nil)
}
