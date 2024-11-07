package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"

	models "github.com/vkupriya/gophkeeper/internal/server/models"
)

const (
	defaultContextTimeout time.Duration = 3 * time.Second
	defaultJWTTokenTTL    time.Duration = 3600 * time.Second
	defaultJWTKey         string        = "vcwYCYkum_2Fsukk"
	defaultAddress        string        = "localhost:8080"
)

func NewConfig() (*models.Config, error) {
	logConfig := zap.NewDevelopmentConfig()
	logger, err := logConfig.Build()
	if err != nil {
		return &models.Config{}, fmt.Errorf("failed to initialize Logger: %w", err)
	}

	a := flag.String("a", defaultAddress, "Gophermart server host address and port.")
	d := flag.String("d", "", "PostgreSQL DSN")

	flag.Parse()

	if *a == defaultAddress {
		if envAddr, ok := os.LookupEnv("RUN_ADDRESS"); ok {
			a = &envAddr
		}
	}

	if *d == "" {
		if envDSN, ok := os.LookupEnv("DATABASE_URI"); ok {
			d = &envDSN
		} else {
			return &models.Config{}, errors.New("postgreSQL DSN is missing")
		}
	}

	var JWTKey string
	if envJWT, ok := os.LookupEnv("JWT"); ok {
		JWTKey = envJWT
	}

	return &models.Config{
		Address:        *a,
		Logger:         logger,
		PostgresDSN:    *d,
		ContextTimeout: defaultContextTimeout,
		JWTKey:         JWTKey,
		JWTTokenTTL:    defaultJWTTokenTTL,
	}, nil
}
