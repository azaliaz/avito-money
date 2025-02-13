package rest

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/azaliaz/avito-money/internal/application"
	"strconv"
	"strings"
)

func (api *Service) Auth(ctx *fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := ctx.BodyParser(&req); err != nil {

		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid jsn",
		})
	}
	res, err := api.app.Auth(ctx.Context(), &application.AuthRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid credentials",
		})
	}

	return ctx.JSON(fiber.Map{
		"token": res.Token,
	})
}

func (api *Service) BuyItem(ctx *fiber.Ctx) error {
	_, err := api.app.BuyItem(ctx.Context(), &application.BuyItemRequest{
		Token: api.getToken(ctx),
		Item:  ctx.Params("item"),
	})
	if err != nil {
		return err
	}
	return nil
}

func (api *Service) Info(ctx *fiber.Ctx) error {
	res, err := api.app.GetInfo(ctx.Context(), &application.GetInfoRequest{
		Token: api.getToken(ctx),
	})
	if err != nil {
		return err
	}

	resInventory := make([]struct {
		// Quantity Количество предметов.
		Quantity *int `json:"quantity,omitempty"`

		// Type Тип предмета.
		Type *string `json:"type,omitempty"`
	}, 0, len(res.Inventory))
	for _, productStock := range res.Inventory {
		resInventory = append(resInventory, struct {
			// Quantity Количество предметов.
			Quantity *int `json:"quantity,omitempty"`

			// Type Тип предмета.
			Type *string `json:"type,omitempty"`
		}{
			Quantity: &productStock.Quantity,
			Type:     &productStock.Type,
		})
	}

	sent := make([]struct {
		// Amount Количество отправленных монет.
		Amount *int `json:"amount,omitempty"`

		// ToUser Имя пользователя, которому отправлены монеты.
		ToUser *string `json:"toUser,omitempty"`
	}, 0, len(res.CoinHistory.Sent))
	for _, tr := range res.CoinHistory.Sent {
		toUser := strconv.FormatUint(tr.ToUser, 10)
		sent = append(sent, struct {
			// Amount Количество отправленных монет.
			Amount *int `json:"amount,omitempty"`

			// ToUser Имя пользователя, которому отправлены монеты.
			ToUser *string `json:"toUser,omitempty"`
		}{
			Amount: &tr.Amount,
			ToUser: &toUser,
		})
	}
	received := make([]struct {
		// Amount Количество полученных монет.
		Amount *int `json:"amount,omitempty"`

		// FromUser Имя пользователя, который отправил монеты.
		FromUser *string `json:"fromUser,omitempty"`
	}, 0, len(res.CoinHistory.Received))
	for _, tr := range res.CoinHistory.Received {
		fromUser := strconv.FormatUint(tr.FromUser, 10)
		received = append(received, struct {
			// Amount Количество полученных монет.
			Amount *int `json:"amount,omitempty"`

			// FromUser Имя пользователя, который отправил монеты.
			FromUser *string `json:"fromUser,omitempty"`
		}{
			Amount:   &tr.Amount,
			FromUser: &fromUser,
		})
	}

	response := struct {
		CoinHistory *struct {
			Received *[]struct {
				// Amount Количество полученных монет.
				Amount *int `json:"amount,omitempty"`

				// FromUser Имя пользователя, который отправил монеты.
				FromUser *string `json:"fromUser,omitempty"`
			} `json:"received,omitempty"`
			Sent *[]struct {
				// Amount Количество отправленных монет.
				Amount *int `json:"amount,omitempty"`

				// ToUser Имя пользователя, которому отправлены монеты.
				ToUser *string `json:"toUser,omitempty"`
			} `json:"sent,omitempty"`
		} `json:"coinHistory,omitempty"`

		// Coins Количество доступных монет.
		Coins     *int `json:"coins,omitempty"`
		Inventory *[]struct {
			// Quantity Количество предметов.
			Quantity *int `json:"quantity,omitempty"`

			// Type Тип предмета.
			Type *string `json:"type,omitempty"`
		} `json:"inventory,omitempty"`
	}{
		CoinHistory: &struct {
			Received *[]struct {
				// Amount Количество полученных монет.
				Amount *int `json:"amount,omitempty"`

				// FromUser Имя пользователя, который отправил монеты.
				FromUser *string `json:"fromUser,omitempty"`
			} `json:"received,omitempty"`
			Sent *[]struct {
				// Amount Количество отправленных монет.
				Amount *int `json:"amount,omitempty"`

				// ToUser Имя пользователя, которому отправлены монеты.
				ToUser *string `json:"toUser,omitempty"`
			} `json:"sent,omitempty"`
		}{Received: &received, Sent: &sent},
		Coins:     &res.Coins,
		Inventory: &resInventory,
	}
	u, err := json.Marshal(response)
	if err != nil {
		return err
	}

	return ctx.SendString(string(u))
}

func (api *Service) SendCoin(ctx *fiber.Ctx) error {
	var req struct {
		Amount string `json:"amount"`
		ToUser string `json:"toUser"`
	}
	if err := ctx.BodyParser(&req); err != nil {

		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid jsn",
		})
	}
	amount, err := strconv.Atoi(req.Amount)
	if err != nil {
		return err
	}
	toUser, err := strconv.ParseUint(req.ToUser, 10, 64)
	if err != nil {
		return err
	}
	_, err = api.app.SendCoin(ctx.Context(), &application.SendCoinRequest{
		Token:  api.getToken(ctx),
		Amount: amount,
		ToUser: toUser,
	})
	if err != nil {
		return err
	}

	return nil
}

func (api *Service) getToken(ctx *fiber.Ctx) string {
	return strings.TrimPrefix(ctx.Get("Authorization"), "Bearer ")
}
