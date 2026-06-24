package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	RedisAddr string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env file didnt find , uses system environment variables ")
	}

	return &Config{
		RedisAddr: os.Getenv("REDIS_ADDR"),
	}
}
