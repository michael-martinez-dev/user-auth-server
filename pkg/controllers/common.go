package controllers

import (
	"github.com/mixedmachine/EfficientLife/user-auth/pkg/repository"
	"github.com/mixedmachine/EfficientLife/user-auth/pkg/util"

	"log"

	"github.com/gofiber/fiber/v2"
)

func AuthRequest(ctx *fiber.Ctx, tokensRepo repository.TokenRepository) (string, error) {
	token := string(ctx.Request().Header.Peek("Authorization"))
	// log.Printf("Token: %s\n", token)
	if token == "" {
		return "", util.ErrInvalidAuthToken
	}
	user, err := tokensRepo.Retrieve(token)
	if user == "" || err != nil {
		log.Printf("User: %s\n", user)
		log.Printf("Error: %s\n", err)
		return "", util.ErrUnauthorized
	}

	return user, nil
}
