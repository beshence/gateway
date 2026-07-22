package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWT struct {
	secret []byte
	ttl    time.Duration
}

type Claims struct {
	BankID string
}

func NewJWTManager(secret string, ttl time.Duration) *JWT {
	return &JWT{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

func (m *JWT) GenerateToken(bankID string) (string, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(m.ttl)

	claims := jwt.MapClaims{
		"sub": bankID,
		"typ": "access",
		"iat": now.Unix(),
		"exp": expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(m.secret)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func (m *JWT) ParseToken(tokenString string) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}

	parsedToken, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return m.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, err
	}

	if !parsedToken.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}

func ClaimsFromToken(claims jwt.MapClaims) (Claims, bool) {
	bankID, bankIDOk := claims["sub"].(string)
	if !bankIDOk {
		return Claims{}, false
	}

	if bankID == "" {
		return Claims{}, false
	}

	return Claims{
		BankID: bankID,
	}, true
}
