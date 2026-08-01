package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
	R2          R2Config
}

type R2Config struct {
	Endpoint        string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	UsePathStyle    bool
}

func Load() (Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	r2 := R2Config{
		Endpoint:        os.Getenv("R2_ENDPOINT"),
		Bucket:          os.Getenv("R2_BUCKET"),
		AccessKeyID:     os.Getenv("R2_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("R2_SECRET_ACCESS_KEY"),
		UsePathStyle:    os.Getenv("R2_USE_PATH_STYLE") == "true",
	}
	if r2.Endpoint == "" || r2.Bucket == "" || r2.AccessKeyID == "" || r2.SecretAccessKey == "" {
		return Config{}, fmt.Errorf("R2 storage configuration is incomplete")
	}

	return Config{
		Port:        port,
		DatabaseURL: databaseURL,
		R2:          r2,
	}, nil
}

func (c Config) Address() string {
	return fmt.Sprintf(":%s", c.Port)
}
