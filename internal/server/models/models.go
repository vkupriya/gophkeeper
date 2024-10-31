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

const (
	UnknownSecret int32 = iota
	TextSecret
	BinarySecret
	CardSecret
	FileSecret
)

func TypeToProto(st string) int32 {
	switch st {
	case "text":
		return TextSecret
	case "binary":
		return BinarySecret
	case "card":
		return CardSecret
	case "file":
		return FileSecret
	case "unknown":
		return UnknownSecret
	default:
		return UnknownSecret
	}
}

func ProtoToType(i int32) string {
	switch i {
	case TextSecret:
		return "text"
	case BinarySecret:
		return "binary"
	case CardSecret:
		return "card"
	case FileSecret:
		return "file"
	case UnknownSecret:
		return "unknown"
	default:
		return "unknown"
	}
}
