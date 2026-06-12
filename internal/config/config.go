package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type Config struct {
	ConnectionUri      string
	JWTSecret          string
	JWTExpirationHours int
	LogLevel           logrus.Level
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	level, err := logrus.ParseLevel(getEnv("LOG_LEVEL", "info"))

	if err != nil {
		level = logrus.InfoLevel
	}

	jwtExpHours, _ := strconv.Atoi(getEnv("JWT_EXPIRATION_HOURS", "24"))

	return &Config{
		ConnectionUri:      getEnv("MONGO_URI", "mongodb://localhost:27017"),
		JWTSecret:          getEnv("JWT_SECRET", "LOLOLOL"),
		JWTExpirationHours: jwtExpHours,
		LogLevel:           level,
	}, err

}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
