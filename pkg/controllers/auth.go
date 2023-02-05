package controllers

import (
	"github.com/mixedmachine/user-auth-server/pkg/models"
	"github.com/mixedmachine/user-auth-server/pkg/repository"
	"github.com/mixedmachine/user-auth-server/pkg/security"
	"github.com/mixedmachine/user-auth-server/pkg/util"

	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/asaskevich/govalidator.v9"
)

// AuthController interface defines the contract for the AuthController
type AuthController interface {
	Ping(ctx *fiber.Ctx) error
	SignUp(ctx *fiber.Ctx) error
	SignIn(ctx *fiber.Ctx) error
	RefreshToken(ctx *fiber.Ctx) error
	Authenticator(ctx *fiber.Ctx) error
}

// authController struct implements the AuthController interface
type authController struct {
	usersRepo  repository.UsersRepository
	tokensRepo repository.TokenRepository
}

// NewAuthController constructs a new instance of AuthController with given repository dependencies
func NewAuthController(repos map[string]interface{}) AuthController {
	return &authController{
		usersRepo:  repos["users"].(repository.UsersRepository),
		tokensRepo: repos["tokens"].(repository.TokenRepository),
	}
}

// Ping Handler Function for Health Check
// @Summary Health Check
// @Description Health Check
// @Tags Status
// @Accept */*
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/v1/ping [get]
func (c *authController) Ping(ctx *fiber.Ctx) error {
	return ctx.
		Status(http.StatusOK).
		JSON(fiber.Map{
			"message": "pong",
		})
}

/********************************************************
 * 		  Handler Functions for Authentication			*
 ********************************************************/

// SignUp Handler Function verifies the user input and creates a new user in the database
// @Summary Sign Up
// @Description Sign Up
// @Tags Auth
// @Accept json
// @Produce json
// @Param name body string true "Name"
// @Param email body string true "Email"
// @Param password body string true "Password"
// @Param admin body bool false "Admin"
// @Success 201 {object} models.User
// @Failure 400 {object} util.JError
// @Failure 422 {object} util.JError
// @Router /api/v1/signup [post]
func (c *authController) SignUp(ctx *fiber.Ctx) error {
	var newUser models.User

	err := ctx.BodyParser(&newUser)
	if err != nil {
		return ctx.
			Status(http.StatusUnprocessableEntity).
			JSON(util.NewJError(err))
	}

	err = verifyUser(&newUser, c)
	if err != nil {
		return ctx.
			Status(http.StatusBadRequest).
			JSON(util.NewJError(err))
	}

	newUser.CreatedAt = time.Now()
	newUser.UpdatedAt = newUser.CreatedAt
	newUser.Id = primitive.NewObjectID()

	err = c.usersRepo.Save(&newUser)
	if err != nil {
		return ctx.
			Status(http.StatusBadRequest).
			JSON(util.NewJError(err))
	}

	return ctx.
		Status(http.StatusCreated).
		JSON(newUser)
}

// SignIn Handler Function verifies the user input and returns a new token
// @Summary Sign In
// @Description Sign In
// @Tags Auth
// @Accept json
// @Produce json
// @Param email body string true "Email"
// @Param password body string true "Password"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} util.JError
// @Failure 422 {object} util.JError
// @Router /api/v1/signin [post]
func (c *authController) SignIn(ctx *fiber.Ctx) error {
	var input models.User
	err := ctx.BodyParser(&input)
	if err != nil {
		return ctx.
			Status(http.StatusUnprocessableEntity).
			JSON(util.NewJError(err))
	}

	input.Email = util.NormalizeEmail(input.Email)
	user, err := c.usersRepo.GetByEmail(input.Email)
	if err != nil {
		log.Printf("c.usersRepo.GetByEmail| %s signin failed: %v\n", input.Email, err.Error())
		return ctx.
			Status(http.StatusUnauthorized).
			JSON(util.NewJError(util.ErrInvalidCredentials))
	}

	err = security.VerifyPassword(user.Password, input.Password)
	if err != nil {
		log.Printf("security.VerifyPassword| %s signin failed: %v\n", input.Email, err.Error())
		return ctx.
			Status(http.StatusUnauthorized).
			JSON(util.NewJError(util.ErrInvalidCredentials))
	}

	token, err := security.NewToken(user.Id.Hex())
	if err != nil {
		log.Printf("security.NewToken| %s signin failed: %v\n", input.Email, err.Error())
		return ctx.
			Status(http.StatusUnauthorized).
			JSON(util.NewJError(err))
	}
	err = c.tokensRepo.Create(token, user.Id.Hex(), true)
	if err != nil {
		log.Printf("c.tokensRepo.Create| %s signin failed: %v\n", input.Email, err.Error())
		return ctx.
			Status(http.StatusUnauthorized).
			JSON(util.NewJError(err))
	}

	return ctx.
		Status(http.StatusOK).
		JSON(fiber.Map{
			"user":  user,
			"token": token,
		})
}

// RefreshToken Handler Function verifies the user input removes old token and returns a new token
// @Summary Refresh Token
// @Description Refresh Token
// @Tags Auth
// @Accept json
// @Produce json
// @Param Authorization header string true "specific user token"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} util.JError
// @Failure 422 {object} util.JError
// @Router /api/v1/refresh [post]
func (c *authController) RefreshToken(ctx *fiber.Ctx) error {
	userId, err := AuthRequest(ctx, c.tokensRepo)
	if err != nil {
		log.Printf("AuthRequest| %s refresh failed: %v\n", userId, err.Error())
		return ctx.
			Status(http.StatusUnauthorized).
			JSON(util.NewJError(err))
	}

	token, err := security.NewToken(userId)
	if err != nil {
		log.Printf("security.NewToken| %s refresh failed: %v\n", userId, err.Error())
		return ctx.
			Status(http.StatusUnauthorized).
			JSON(util.NewJError(err))
	}

	err = c.tokensRepo.Create(token, userId, true)
	if err != nil {
		log.Printf("c.tokensRepo.Create| %s refresh failed: %v\n", userId, err.Error())
		return ctx.
			Status(http.StatusUnauthorized).
			JSON(util.NewJError(err))
	}

	err = c.tokensRepo.Delete(string(ctx.Request().Header.Peek("Authorization")))
	if err != nil {
		log.Printf("c.tokensRepo.Delete| %s refresh failed: %v\n", userId, err.Error())
		return ctx.
			Status(http.StatusUnauthorized).
			JSON(util.NewJError(err))
	}

	return ctx.
		Status(http.StatusOK).
		JSON(fiber.Map{
			"token": token,
		})
}

// Authenticator Handler Function takes the token from the request header and returns the user id
// associated with the token
// @Summary Authenticator
// @Description Authenticator
// @Tags Auth
// @Accept json
// @Produce json
// @Param Authorization header string true "specific user token"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} util.JError
// @Failure 422 {object} util.JError
// @Router /api/v1/auth [post]
func (c *authController) Authenticator(ctx *fiber.Ctx) error {
	userId, err := AuthRequest(ctx, c.tokensRepo)
	if err != nil {
		return ctx.
			Status(http.StatusUnauthorized).
			JSON(util.NewJError(err))
	}
	return ctx.
		Status(http.StatusOK).
		JSON(fiber.Map{
			"user_id": userId,
		})
}

/********************************************************
* 					Helper functions					*
*********************************************************/

// verifyUser verifies the user input and returns an error if the input is invalid
func verifyUser(user *models.User, c *authController) error {
	if user == nil {
		return util.ErrEmptyUser
	}
	if user.Name == "" {
		return util.ErrEmptyName
	}
	if user.Email == "" {
		return util.ErrInvalidEmail
	}
	if user.Password == "" {
		return util.ErrEmptyPassword
	}

	user.Email = util.NormalizeEmail(user.Email)
	if !govalidator.IsEmail(user.Email) {
		return util.ErrInvalidEmail
	}

	exists, err := c.usersRepo.GetByEmail(user.Email)
	if err != mongo.ErrNoDocuments {
		return err
	}

	if exists != nil {
		err = util.ErrEmailAlreadyExists
		return err
	}

	if strings.TrimSpace(user.Password) == "" {
		return util.ErrEmptyPassword
	}

	user.Password, err = security.EncryptPassword(user.Password)
	if err != nil {
		return err
	}

	return nil
}
