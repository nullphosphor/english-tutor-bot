package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Token           string
	AdminTelegramID int64
	DatabaseURL     string
	TimeZone        *time.Location
}

func LoadConfig() (Config, error) {
	var config Config

	if err := godotenv.Load(".env"); err != nil {
		return Config{}, fmt.Errorf("cannot load config: %w", err)
	}

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if strings.TrimSpace(botToken) == "" {
		return Config{}, errors.New("empty bot token")
	}

	id := os.Getenv("ADMIN_TELEGRAM_ID")
	if strings.TrimSpace(id) == "" {
		return Config{}, errors.New("empty adminTelegramID")

	}

	telegramAdminID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return Config{}, fmt.Errorf("parse id: %w", err)
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(databaseURL) == "" {
		return Config{}, errors.New("empty databaseURL")
	}

	tz := os.Getenv("TIME_ZONE")
	if tz == "" {
		tz = "UTC"
	}

	zone, err := time.LoadLocation(tz)
	if err != nil {
		return Config{}, fmt.Errorf("load location: %w", err)
	}

	config.TimeZone = zone
	config.Token = botToken
	config.AdminTelegramID = telegramAdminID
	config.DatabaseURL = databaseURL

	return config, nil
}

func LoadConfigMust() Config {
	config, err := LoadConfig()

	if err != nil {
		panic(fmt.Errorf("load config must: %w", err))
	}

	return config
}
