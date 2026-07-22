package api

import "gateway/internal/auth"

type Dependencies struct {
	JWTManager *auth.JWT
}

func NewDependencies(
	jwt *auth.JWT,
) *Dependencies {
	return &Dependencies{
		JWTManager: jwt,
	}
}
