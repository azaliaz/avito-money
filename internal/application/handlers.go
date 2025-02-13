package application

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/azaliaz/avito-money/internal/storage"
)

func (s *Service) Auth(ctx context.Context, request *AuthRequest) (*AuthResponse, error) {
	
	if request.Password == "" {
        return nil, errors.New("password cannot be empty")
    }
	res, err := s.db.Auth(ctx, &storage.AuthRequest{
		UserName: request.Username,
		PassHash: request.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("error auth in db: %w", err)
	}
	if res.PassHash != request.Password {
		return nil, errors.New("invalid creds")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": res.UserId,
	})
	t, err := token.SignedString([]byte(s.config.Secret))
	if err != nil {
		
		return nil, fmt.Errorf("error sign token: %w", err)
	}
	return &AuthResponse{
		Token: t,
	}, nil

}
func (s *Service) GetInfo(ctx context.Context, request *GetInfoRequest) (*GetInfoResponse, error) {
	userId, err := s.userIdFromToken(request.Token)
	if err != nil {
		return nil, err
	}

	inventory, err := s.db.GetInventory(ctx, userId)
	if err != nil {
		return nil, err
	}
	resInventory := make([]*ProductStock, 0, len(inventory))
	for _, productStock := range inventory {
		resInventory = append(resInventory, &ProductStock{
			Type:     productStock.Type,
			Quantity: productStock.Quantity,
		})
	}
	balance, err := s.db.GetBalance(ctx, userId)
	if err != nil {
		return nil, err
	}
	coinHistory, err := s.db.GetCoinHistory(ctx, userId)
	if err != nil {
		return nil, err
	}
	received := make([]*Transaction, 0, len(coinHistory.Received))
	for _, transaction := range coinHistory.Received {
		received = append(received, &Transaction{
			Amount:    transaction.Amount,
			FromUser:  transaction.FromUser,
			ToUser:    transaction.ToUser,
			CreatedAt: transaction.CreatedAt,
		})
	}
	sent := make([]*Transaction, 0, len(coinHistory.Sent))
	for _, transaction := range coinHistory.Sent {
		sent = append(sent, &Transaction{
			Amount:    transaction.Amount,
			FromUser:  transaction.FromUser,
			ToUser:    transaction.ToUser,
			CreatedAt: transaction.CreatedAt,
		})
	}
	return &GetInfoResponse{
		CoinHistory: &CoinHistory{
			Received: received,
			Sent:     sent,
		},
		Coins:     balance,
		Inventory: resInventory,
	}, nil
}
func (s *Service) SendCoin(ctx context.Context, request *SendCoinRequest) (*SendCoinResponse, error) {
	userId, err := s.userIdFromToken(request.Token)
	if err != nil {
		return nil, err
	}

	_, err = s.db.SendCoin(ctx, &storage.SendCoinRequest{
		UserId: userId,
		Amount: request.Amount,
		ToUser: request.ToUser,
	})
	if err != nil {
		return nil, err
	}
	return &SendCoinResponse{}, nil
}
func (s *Service) BuyItem(ctx context.Context, request *BuyItemRequest) (*BuyItemResponse, error) {
	userId, err := s.userIdFromToken(request.Token)
	if err != nil {
		return nil, err
	}
	_, err = s.db.BuyItem(ctx, &storage.BuyItemRequest{
		UserId: userId,
		Item:   request.Item,
	})
	if err != nil {
		return nil, err
	}
	return &BuyItemResponse{}, nil
}

func (s *Service) userIdFromToken(token string) (uint64, error) {
	claims := jwt.MapClaims{}
	t, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.Secret), nil
	})
	if err != nil {
		return 0, err
	}
	if !t.Valid {
		return 0, fmt.Errorf("invalid token")
	}
	userId := uint64(claims["user_id"].(float64))
	return userId, nil
}
