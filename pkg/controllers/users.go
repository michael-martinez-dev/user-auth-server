package controllers

import (
	"github.com/mixedmachine/user-auth-server/pkg/models"
	"github.com/mixedmachine/user-auth-server/pkg/repository"
	"github.com/mixedmachine/user-auth-server/pkg/security"
	"github.com/mixedmachine/user-auth-server/pkg/util"

	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/asaskevich/govalidator.v9"
)

// UserController defines the interface for user controller
type UserController interface {
	GetUser(ctx *fiber.Ctx) error
	GetUsers(ctx *fiber.Ctx) error
	PutUser(ctx *fiber.Ctx) error
	DeleteUser(ctx *fiber.Ctx) error
}

// userController implements UserController
type userController struct {
	usersRepo  repository.UsersRepository
	tokensRepo repository.TokenRepository
}

// NewUserController constructs a new instance of UserController with given repository dependencies
func NewUserController(repos map[string]interface{}) UserController {
	return &userController{
		usersRepo:  repos["users"].(repository.UsersRepository),
		tokensRepo: repos["tokens"].(repository.TokenRepository),
	}
}

/********************************************************
 *				Handler Functions for Users				*
 ********************************************************/

// GetUser returns a user by id
// @Summary Get a user by id
// @Description Get a user by id
// @Tags users
// @Accept  json
// @Produce  json
// @Param id path string true "User ID"
// @Param Authorization header string true "specific user token"
// @Success 200 {object} models.User
// @Failure 400 {object} util.JError
// @Failure 401 {object} util.JError
// @Failure 500 {object} util.JError
// @Router /api/v1/users/{id} [get]
func (c *userController) GetUser(ctx *fiber.Ctx) error {
	userId, err := AuthRequest(ctx, c.tokensRepo)
	if err != nil {
		return ctx.
			Status(http.StatusUnauthorized).
			JSON(util.NewJError(err))
	}
	user, err := c.usersRepo.GetById(userId)
	if err != nil {
		return ctx.
			Status(http.StatusInternalServerError).
			JSON(util.NewJError(err))
	}
	return ctx.
		Status(http.StatusOK).
		JSON(user)
}

// GetUsers returns all users
// @Summary Get all users
// @Description Get all users
// @Tags users
// @Accept  json
// @Produce  json
// @Success 200 {array} models.User
// @Failure 400 {object} util.JError
// @Failure 401 {object} util.JError
// @Failure 500 {object} util.JError
// @Router /api/v1/users [get]
func (c *userController) GetUsers(ctx *fiber.Ctx) error {
	users, err := c.usersRepo.GetAll()
	if err != nil {
		return ctx.
			Status(http.StatusInternalServerError).
			JSON(util.NewJError(err))
	}
	return ctx.
		Status(http.StatusOK).
		JSON(users)
}

// PutUser updates a user by id
// @Summary Update a user by id
// @Description Update a user by id
// @Tags users
// @Accept  json
// @Produce  json
// @Param id path string true "User ID"
// @Param email body string true "User email"
// @Param name body string true "User name"
// @Param password body string true "User password"
// @Param Authorization header string true "specific user token"
// @Success 200 {object} models.User
// @Failure 400 {object} util.JError
// @Failure 401 {object} util.JError
// @Failure 422 {object} util.JError
// @Router /api/v1/users/{id} [put]
func (c *userController) PutUser(ctx *fiber.Ctx) error {
	userId, err := AuthRequest(ctx, c.tokensRepo)
	if err != nil {
		return ctx.
			Status(http.StatusUnauthorized).
			JSON(util.NewJError(err))
	}
	var update models.User
	err = ctx.BodyParser(&update)
	if err != nil {
		return ctx.
			Status(http.StatusUnprocessableEntity).
			JSON(util.NewJError(err))
	}
	update.Email = util.NormalizeEmail(update.Email)
	if !govalidator.IsEmail(update.Email) {
		return ctx.
			Status(http.StatusBadRequest).
			JSON(util.NewJError(util.ErrInvalidEmail))
	}
	exists, err := c.usersRepo.GetByEmail(update.Email)
	if err == mongo.ErrNoDocuments || exists.Id.Hex() == userId {
		user, err := c.usersRepo.GetById(userId)
		if err != nil {
			return ctx.
				Status(http.StatusBadRequest).
				JSON(util.NewJError(err))
		}
		if update.Name != "" {
			user.Name = update.Name
		}
		if update.Email != "" {
			user.Email = update.Email
		}
		if update.Password != "" {
			update.Password, err = security.EncryptPassword(update.Password)
			if err != nil {
				return ctx.
					Status(http.StatusBadRequest).
					JSON(util.NewJError(err))
			}
			user.Password = update.Password
		}
		user.UpdatedAt = time.Now()
		err = c.usersRepo.Update(user)
		if err != nil {
			return ctx.
				Status(http.StatusUnprocessableEntity).
				JSON(util.NewJError(err))
		}
		return ctx.
			Status(http.StatusOK).
			JSON(user)
	}

	if exists != nil {
		err = util.ErrEmailAlreadyExists
	}

	return ctx.
		Status(http.StatusBadRequest).
		JSON(util.NewJError(err))
}

// DeleteUser deletes a user by id
// @Summary Delete a user by id
// @Description Delete a user by id
// @Tags users
// @Accept  json
// @Produce  json
// @Param id path string true "User ID"
// @Param Authorization header string true "specific user token"
// @Success 204
// @Failure 400 {object} util.JError
// @Failure 401 {object} util.JError
// @Failure 500 {object} util.JError
// @Router /api/v1/users/{id} [delete]
func (c *userController) DeleteUser(ctx *fiber.Ctx) error {
	userId, err := AuthRequest(ctx, c.tokensRepo)
	if err != nil {
		return ctx.
			Status(http.StatusUnauthorized).
			JSON(util.NewJError(err))
	}
	err = c.usersRepo.Delete(userId)
	if err != nil {
		return ctx.
			Status(http.StatusInternalServerError).
			JSON(util.NewJError(err))
	}
	err = c.tokensRepo.Delete(string(ctx.Request().Header.Peek("Authorization")))
	if err != nil {
		return ctx.
			Status(http.StatusInternalServerError).
			JSON(util.NewJError(err))
	}
	ctx.Set("Entity", userId)
	return ctx.SendStatus(http.StatusNoContent)
}
