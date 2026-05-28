package config

import (
	"os"
	"strconv"
)

type Config struct {
	AppPort string
	AppEnv  string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int

	JWTSecret               string
	JWTAccessExpiresMinutes int
	JWTRefreshExpiresDays   int
}

func Load() *Config {
	redisDB, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	jwtAccessExpires, _ := strconv.Atoi(os.Getenv("JWT_ACCESS_EXPIRES_MINUTES"))
	jwtRefreshExpires, _ := strconv.Atoi(os.Getenv("JWT_REFRESH_EXPIRES_DAYS"))

	return &Config{
		AppPort: os.Getenv("APP_PORT"),
		AppEnv:  os.Getenv("APP_ENV"),

		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),

		RedisHost:     os.Getenv("REDIS_HOST"),
		RedisPort:     os.Getenv("REDIS_PORT"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB:       redisDB,

		JWTSecret:               os.Getenv("JWT_SECRET"),
		JWTAccessExpiresMinutes: jwtAccessExpires,
		JWTRefreshExpiresDays:   jwtRefreshExpires,
	}
}
