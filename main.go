package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"key-server/database"
	"key-server/logger"
	"key-server/middleware"
	"key-server/router"
	"os"
)

func init() {
	// Loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log := logger.GetForFile("startup-errors")
		log.Error("No .env file found", zap.Error(err))
	}
	// redis.InitRedis()
}

func main() {
	app := fiber.New(fiber.Config{
		CaseSensitive: true,
		StrictRouting: true,
		AppName:       "Toffee Key Server",
	})

	//app.Use(cors.New())

	app.Use(cors.New(cors.Config{
		AllowHeaders:     "Origin,Content-Type,Accept,Content-Length,Accept-Language,Accept-Encoding,Connection,Access-Control-Allow-Origin,Authorization",
		AllowOrigins:     "*",
		AllowCredentials: false,
		AllowMethods:     "GET,POST,HEAD,PUT,PATCH,OPTIONS",
	}))

	app.Use(helmet.New())
	//app.Use(csrf.New())
	app.Use(middleware.RecoverFromPanic)

	database.ConnectDB()

	/*if err := redis.PingRedis(); err != nil {
		// Handle error if Redis connection fails
		log := logger.GetForFile("startup-errors")
		log.Error("Failed to ping Redis client", zap.Error(err))
	}

	defer func() {
		err := redis.CloseRedisClient()
		if err != nil {
			log := logger.GetForFile("startup-errors")
			log.Error("Failed to close Redis client", zap.Error(err))
		}
	}()*/

	router.SetupRoutes(app)

	if err := app.Listen(":" + os.Getenv("APP_PORT")); err != nil {
		log := logger.GetForFile("startup-errors")
		log.Error("Failed to start server", zap.Error(err))
	}

}
