package routes

import (
	_ "github.com/mixedmachine/user-auth-server/api"
	"github.com/mixedmachine/user-auth-server/pkg/controllers"

	"fmt"
	"net/http"

	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

const apiVersion = "v1"

type authRoutes struct {
	authController controllers.AuthController
	userController controllers.UserController
}

func NewAuthRoutes(authController controllers.AuthController, userController controllers.UserController) Routes {
	return &authRoutes{
		authController: authController,
		userController: userController,
	}
}

func (r *authRoutes) Install(app *fiber.App) {
	app.Get("/", serviceInfo)
	app.Get("/swagger/*", swagger.HandlerDefault)
	api := app.Group(fmt.Sprintf("/api/%s", apiVersion))

	// Health check
	api.Get("/ping", r.authController.Ping)

	// Authentication
	api.Post("/signup", r.authController.SignUp)
	api.Post("/signin", r.authController.SignIn)
	api.Post("/refresh", r.authController.RefreshToken)
	api.Get("/auth", r.authController.Authenticator)

	// Users management
	usersGroup := api.Group("/users")
	usersGroup.Get("/", r.userController.GetUsers)
	usersGroup.Get("/:id", r.userController.GetUser)
	usersGroup.Put("/:id", r.userController.PutUser)
	usersGroup.Delete("/:id", r.userController.DeleteUser)
}

// Service info
// @Summary Service info
// @Description Service info
// @Tags Status
// @Accept */*
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router / [get]
func serviceInfo(ctx *fiber.Ctx) error {
	return ctx.
		Status(http.StatusOK).
		JSON(bson.D{
			{Key: "message", Value: "Welcome to EfficientLife"},
			{Key: "service", Value: "user-auth"},
			{Key: "author", Value: "MixedMachine"},
			{Key: "status", Value: http.StatusOK},
			{Key: "version", Value: apiVersion},
			{Key: "api_base_endpoint", Value: "/api/" + apiVersion},
			{Key: "api_endpoints", Value: map[string]string{
				"GET| /":                  "Service info",
				"GET| <api>/ping":         "Health check",
				"POST| <api>/signup":      "Create a new user",
				"POST| <api>/signin":      "Sign in and get token",
				"POST| <api>/refresh":     "Refresh token",
				"GET| <api>/auth":         "Get user based on token",
				"GET| <api>/users/":       "Get all users",
				"GET| <api>/users/:id":    "Get user by id",
				"PUT| <api>/users/:id":    "Update user by id",
				"DELETE| <api>/users/:id": "Delete user by id",
			}},
		})
}
