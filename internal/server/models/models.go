package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type Config struct {
	Logger         *zap.Logger
	Address        string
	PostgresDSN    string
	JWTKey         string
	JWTTokenTTL    time.Duration
	ContextTimeout time.Duration
}

type User struct {
	UserID   string `json:"login"`
	Password string `json:"password"`
}

type Claims struct {
	UserID string
	jwt.RegisteredClaims
}

type Secret struct {
	UserID  string
	Name    string
	Type    string
	Meta    string
	Data    []byte
	Version int64
}

type SecretList []SecretItem

type SecretItem struct {
	Name    string
	Type    string
	Version int64
}
