package api

import (
	"gateway/internal/auth"
	"gateway/internal/signal"
)

type Dependencies struct {
	JWTManager    *auth.JWT
	SignalManager *signal.Manager
}

func NewDependencies(
	jwt *auth.JWT,
	signal *signal.Manager,
) *Dependencies {
	return &Dependencies{
		JWTManager:    jwt,
		SignalManager: signal,
	}
}
