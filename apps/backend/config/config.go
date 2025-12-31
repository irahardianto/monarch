package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port int
	Env  string
	DB   string
}

func Load() (*Config, error) {
	portStr := os.Getenv("MONARCH_PORT")
	port := 9090
	if portStr != "" {
		var err error
		port, err = strconv.Atoi(portStr)
		if err != nil {
			return nil, err
		}
	}

	return &Config{
		Port: port,
		Env:  getEnv("MONARCH_ENV", "development"),
		DB:   os.Getenv("DATABASE_URL"),
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
