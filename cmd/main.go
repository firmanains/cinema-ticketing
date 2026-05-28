package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	"github.com/firmanains/cinema-ticketing/config"
	"github.com/firmanains/cinema-ticketing/internal/handler"
	"github.com/firmanains/cinema-ticketing/internal/repository"
	"github.com/firmanains/cinema-ticketing/internal/service"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using environment variables")
	}

	cfg := config.Load()

	db, err := config.NewDB(cfg)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer db.Close()

	rdb, err := config.NewRedis(cfg)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer rdb.Close()

	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewRefreshTokenRepository(db)
	userSvc := service.NewUserService(userRepo, tokenRepo, cfg, rdb)

	authHandler := handler.NewAuthHandler(userSvc)

	app := fiber.New()

	api := app.Group("/api/v1")
	api.Post("/auth/register", authHandler.Register)
	api.Post("/auth/login", authHandler.Login)

	log.Printf("starting server on port %s", cfg.AppPort)
	if err := app.Listen(":" + cfg.AppPort); err != nil {
		log.Fatalf("server: %v", err)
	}
}
