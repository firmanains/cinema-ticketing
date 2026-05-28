package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	"github.com/firmanains/cinema-ticketing/config"
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

	_ = db
	_ = rdb

	app := fiber.New()

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	log.Printf("starting server on port %s", cfg.AppPort)
	if err := app.Listen(":" + cfg.AppPort); err != nil {
		log.Fatalf("server: %v", err)
	}
}
