package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env				 string
	Port			 string
	DBURL			 string
	StripeSigningKey string
	GracefulTimeout  time.Duration
}

func Load() Config {
	_ = godotenv.Load()

	cfg := Config{
		Env:              getEnv("APP_ENV", "local"),
		Port:             getEnv("PORT", "8080"),
		DBURL:            getEnv("DB_URL", "postgres://postgres:postgres@localhost:5432/webhooks?sslmode=disable"),
		StripeSigningKey: getEnv("STRIPE_SIGNING_SECRET", ""),
		GracefulTimeout:  durationSeconds("GRACEFUL_TIMEOUT_SECONDS", 10),
	}
	return cfg
}

func getEnv(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}

func durationSeconds(key string, def int) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return time.Duration(def) * time.Second
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		log.Printf("invalid int for %s: %v", key, err)
		return time.Duration(def) * time.Second
	}
	return time.Duration(n) * time.Second
}