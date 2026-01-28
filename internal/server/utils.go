package server

import (
	"context"
	"fmt"

	"unipilot/internal/errors"
	"unipilot/internal/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func GetDB(ctx context.Context) (*gorm.DB, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return nil, errors.WrapServer(fmt.Errorf("db not found"), errors.ContextInvalid, "DB not found in context", fiber.StatusInternalServerError)
	}
	return db, nil
}

func GetUser(ctx context.Context) (*models.User, error) {
	user, ok := ctx.Value("user").(*models.User)
	if !ok {
		LogDebug(ctx, "user", user)
		return nil, errors.WrapServer(fmt.Errorf("user not found"), errors.ContextInvalid, "User not found in context", fiber.StatusInternalServerError)
	}
	return user, nil
}
