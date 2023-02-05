package api

import (
	"github.com/mixedmachine/EfficientLife/user-auth/pkg/controllers"
	"github.com/mixedmachine/EfficientLife/user-auth/pkg/db"
	"github.com/mixedmachine/EfficientLife/user-auth/pkg/repository"
	"github.com/mixedmachine/EfficientLife/user-auth/pkg/routes"

	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	
)

func Init() {
	err := godotenv.Load()
	if err != nil {
		log.Panicln(err)
	}
}

func RunUserAuthApiServer() {
	mConn := db.NewMongoConnection()
	rConn := db.NewRedisConnection()
	defer mConn.Close()
	defer rConn.Close()

	app := fiber.New(
		fiber.Config{
			CaseSensitive: true,
			ErrorHandler: func(c *fiber.Ctx, err error) error {
				code := fiber.StatusInternalServerError
				if e, ok := err.(*fiber.Error); ok {
					code = e.Code
				}
				return c.Status(code).JSON(fiber.Map{
					"success": false,
					"error":   err.Error(),
				})
			},
		},
	)
	app.Use(cors.New())
	app.Use(logBuilder())
	app.Use(func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")
		c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
		return c.Next()
	})

	userRepo := repository.NewUserRepository(mConn)
	tokenRepo := repository.NewTokenRepository(rConn)
	repos := map[string]interface{}{
		"users":  userRepo,
		"tokens": tokenRepo,
	}
	authController := controllers.NewAuthController(repos)
	userController := controllers.NewUserController(repos)

	authRoutes := routes.NewAuthRoutes(authController, userController)
	authRoutes.Install(app)

	run(app)
}

func run(app *fiber.App) {
	go func() {
		port := os.Getenv("PORT")
		if port == "" { port = "9090" }
		err := app.Listen(":" + port)
		if err != nil {
			log.Fatal("Could not run server: ",err)
		}
	}()
	sigChn := make(chan os.Signal, 1)
	signal.Notify(sigChn, syscall.SIGINT, syscall.SIGTERM)
	for {
		multiSignalHandler(<-sigChn)
	}
}

func multiSignalHandler(sig os.Signal) {
	switch sig {
	case syscall.SIGINT:
		log.Println("Shutting down gracefully...")
		os.Exit(0)
	case syscall.SIGTERM:
		log.Println("Signal:", sig.String())
		log.Println("Process is killed.")
		os.Exit(0)
	default:
		log.Println("Unhandled/unknown signal")
	}
}

func logBuilder() func(*fiber.Ctx) error {
	var logOutput io.Writer
	if os.Getenv("ENV") == "production" {
		logOutput, _ = os.OpenFile("/var/logs/api.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	} else {
		logOutput = os.Stdout
	}

	return logger.New(
		logger.Config{
			Format:     "${time} ${status} - ${latency} ${method} ${path}",
			TimeFormat: "20060102",
			TimeZone:   "US/Mountain",
			Output:     logOutput,
		},
	)
}