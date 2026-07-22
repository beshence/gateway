package env

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

var (
	ErrJWTSecretRequired = errors.New("JWT_SECRET is required")
	ErrJWTTTLRequired    = errors.New("ACCESS_JWT_TTL_SECONDS is required")
	ErrJWTTTLInvalid     = errors.New("JWT_TTL_SECONDS must be a positive integer")
)

type Env struct {
	JWTSecret     string
	JWTTTLSeconds time.Duration
}

func Load() (Env, error) {
	_ = godotenv.Load()

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return Env{}, ErrJWTSecretRequired
	}
	if jwtSecret == "change_me_to_a_long_random_secret" {
		return Env{}, ErrJWTSecretRequired
	}

	JWTTTLRaw := os.Getenv("JWT_TTL_SECONDS")
	if JWTTTLRaw == "" {
		return Env{}, ErrJWTTTLRequired
	}

	JWTTTLSeconds, err := strconv.Atoi(JWTTTLRaw)
	if err != nil || JWTTTLSeconds <= 0 {
		return Env{}, ErrJWTTTLInvalid
	}

	return Env{
		JWTSecret:     jwtSecret,
		JWTTTLSeconds: time.Duration(JWTTTLSeconds) * time.Second,
	}, nil
}
