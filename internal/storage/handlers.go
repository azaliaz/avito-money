package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
)

func (r *Service) Auth(ctx context.Context, request *AuthRequest) (*AuthResponse, error) {
	conn, err := r.Pool().Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			r.logger.Error("rollback error", err)
		}
	}()
	_, err = tx.Exec(ctx,
		`INSERT INTO users(username, password_hash)
					SELECT @username, @password_hash
					WHERE NOT EXISTS(SELECT 1 FROM users WHERE username = @username)`,
		pgx.NamedArgs{
			"username":      request.UserName,
			"password_hash": request.PassHash,
		},
	)

	var userId uint64
	var userName string
	var passwordHash string
	err = tx.QueryRow(ctx,
		`SELECT id, username, password_hash
				FROM users
				WHERE username = @username`,
		pgx.NamedArgs{
			"username": request.UserName,
		},
	).Scan(&userId, &userName, &passwordHash)
	_, err = tx.Exec(ctx,
		`INSERT INTO coins(user_id, balance)
					SELECT @user_id, 1000
					WHERE NOT EXISTS(SELECT 1 FROM coins WHERE user_id = @user_id)`,
		pgx.NamedArgs{
			"user_id": userId,
		},
	)
	return &AuthResponse{
		UserId:   userId,
		UserName: userName,
		PassHash: passwordHash,
	}, tx.Commit(ctx)
}

func (r *Service) GetInventory(ctx context.Context, userId uint64) ([]*ProductStock, error) {
	conn, err := r.Pool().Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()
	rows, err := conn.Query(ctx,
		`SELECT item, quantity 
			FROM inventory
			WHERE user_id = @user_id`,
		pgx.NamedArgs{
			"user_id": userId,
		},
	)
	if err != nil {
		return nil, err
	}

	var inventory []*ProductStock
	for rows.Next() {
		var productStock ProductStock
		err := rows.Scan(productStock.Type, productStock.Quantity)
		if err != nil {
			return nil, err
		}
		inventory = append(inventory, &productStock)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("cannot fetch user inventory", "err", err)
	}
	return inventory, nil
}

func (r *Service) GetBalance(ctx context.Context, userId uint64) (balance int, err error) {
	conn, err := r.Pool().Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()
	err = conn.QueryRow(ctx,
		`SELECT balance
					FROM coins
					WHERE user_id = @user_id`,
		pgx.NamedArgs{
			"user_id": userId,
		},
	).Scan(&balance)
	if err != nil {
		return 0, err
	}
	return balance, err
}

func (r *Service) GetCoinHistory(ctx context.Context, userId uint64) (*CoinHistory, error) {
	conn, err := r.Pool().Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()
	rows, err := conn.Query(ctx,
		`SELECT from_user_id, to_user_id, amount, created_at
			FROM transactions
			WHERE from_user_id = @user_id`,
		pgx.NamedArgs{
			"user_id": userId,
		},
	)
	if err != nil {
		return nil, err
	}
	var sent []*Transaction
	for rows.Next() {
		var transaction Transaction
		err := rows.Scan(&transaction.FromUser, &transaction.ToUser, &transaction.Amount, &transaction.CreatedAt)
		if err != nil {
			return nil, err
		}
		sent = append(sent, &transaction)
	}
	if err := rows.Err(); err != nil {
		r.logger.Error("cannot fetch user inventory", "err", err)
	}

	rows, err = conn.Query(ctx,
		`SELECT from_user_id, to_user_id, amount, created_at
			FROM transactions
			WHERE to_user_id = @user_id`,
		pgx.NamedArgs{
			"user_id": userId,
		},
	)
	if err != nil {
		return nil, err
	}
	var received []*Transaction
	for rows.Next() {
		var transaction Transaction
		err := rows.Scan(&transaction.FromUser, &transaction.ToUser, &transaction.Amount, &transaction.CreatedAt)
		if err != nil {
			return nil, err
		}
		received = append(received, &transaction)
	}
	if err := rows.Err(); err != nil {
		r.logger.Error("cannot fetch user inventory", "err", err)
	}

	return &CoinHistory{
		Sent:     sent,
		Received: received,
	}, nil
}

func (r *Service) SendCoin(ctx context.Context, request *SendCoinRequest) (*SendCoinResponse, error) {
	conn, err := r.Pool().Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			r.logger.Error("rollback error", err)
		}
	}()
	res, err := tx.Exec(ctx,
		`UPDATE coins
				SET balance = balance - @amount
				WHERE user_id = @user_id`,
		pgx.NamedArgs{
			"amount":  request.Amount,
			"user_id": request.UserId,
		},
	)
	if err != nil {
		return nil, err
	}
	if res.RowsAffected() == 0 {
		return nil, fmt.Errorf("current user not found")
	}
	res, err = tx.Exec(ctx,
		`UPDATE coins
					SET balance = balance + @amount
					WHERE user_id = @user_id`,
		pgx.NamedArgs{
			"amount":  request.Amount,
			"user_id": request.ToUser,
		},
	)
	if err != nil {
		return nil, err
	}
	if res.RowsAffected() == 0 {
		return nil, fmt.Errorf("target user not found")
	}
	_, err = tx.Exec(ctx,
		`INSERT INTO transactions (from_user_id, to_user_id, amount)
					VALUES (@from_user_id, @to_user_id, @amount)`,
		pgx.NamedArgs{
			"from_user_id": request.UserId,
			"to_user_id":   request.ToUser,
			"amount":       request.Amount,
		},
	)
	if err != nil {
		return nil, err
	}
	return nil, tx.Commit(ctx)
}
func (r *Service) BuyItem(ctx context.Context, request *BuyItemRequest) (*BuyItemResponse, error) {
	conn, err := r.Pool().Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			r.logger.Error("rollabck error", err)
		}
	}()
	var amount int
	err = tx.QueryRow(ctx,
		`SELECT amount
				FROM items
				WHERE name = @item`,
		pgx.NamedArgs{
			"item": request.Item,
		},
	).Scan(&amount)
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(ctx,
		`UPDATE coins SET balance = balance - @amount`,
		pgx.NamedArgs{
			"amount": amount,
		},
	)
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(ctx,
		`UPDATE inventory SET quantity = quantity + 1
					WHERE user_id = @user_id AND item = @item`,
		pgx.NamedArgs{
			"user_id": request.UserId,
			"item":    request.Item,
		},
	)
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(ctx,
		`INSERT INTO inventory(user_id, item, quantity)
				SELECT @user_id, @item, 1
				WHERE NOT EXISTS (SELECT 1 FROM inventory WHERE user_id = @user_id AND item = @item)`,
		pgx.NamedArgs{
			"user_id": request.UserId,
			"item":    request.Item,
		},
	)
	if err != nil {
		return nil, err
	}
	return nil, tx.Commit(ctx)
}