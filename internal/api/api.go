package api

import (
	"gateway/internal/auth"
	"gateway/internal/websocket"
)

type Dependencies struct {
	JWTManager       *auth.JWT
	WebSocketManager *websocket.Manager
}

func NewDependencies(
	jwt *auth.JWT,
	ws *websocket.Manager,
) *Dependencies {
	return &Dependencies{
		JWTManager:       jwt,
		WebSocketManager: ws,
	}
}
