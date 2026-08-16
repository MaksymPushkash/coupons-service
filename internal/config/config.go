package config

import "os"

type Config struct {
	HTTPAddress string
	DatabaseURL string
}

func Load() Config {
	return Config{
		HTTPAddress: getEnv("HTTP_ADDRESS", ":8080"),
		DatabaseURL: getEnv(
			"DATABASE_URL",
			"postgres://postgres:postgres@localhost:5432/coupons?sslmode=disable",
		),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}
	return value
}
