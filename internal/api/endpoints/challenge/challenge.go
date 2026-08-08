package challenge

import (
	"crypto/rand"
	"encoding/base64"
	"gateway/internal/api"
	"gateway/internal/memory"
	"net/http"
	"time"

	"github.com/cloudflare/circl/sign/slhdsa"
	"github.com/gin-gonic/gin"
)

func GetChallengeV1() gin.HandlerFunc {
	return func(c *gin.Context) {
		bankID := c.Param("bankId")

		memory.Mutex.Lock()

		_, ok := memory.Banks[bankID]

		memory.Mutex.Unlock()

		if !ok {
			c.JSON(http.StatusNotFound, gin.H{
				"err":    "NO_BANK",
				"errmsg": "we don't have information about this bank; first send public key via POST /api/bank/{bankID}/pk",
			})
			return
		}

		memory.Mutex.Lock()

		challenge, ok := memory.Challenges[bankID]

		memory.Mutex.Unlock()

		if ok {
			if challenge.ExpiresAt.After(time.Now()) {
				c.JSON(http.StatusOK, gin.H{
					"err":   "0",
					"nonce": base64.RawURLEncoding.EncodeToString(challenge.Nonce),
				})
				return
			}
		}

		nonce := make([]byte, 64)

		_, err := rand.Read(nonce)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"err":    "INTERNAL_ERROR",
				"errmsg": "failed to generate nonce",
			})
			return
		}

		memory.Mutex.Lock()

		memory.Challenges[bankID] = memory.Challenge{
			Nonce:     nonce,
			ExpiresAt: time.Now().Add(time.Minute),
		}

		memory.Mutex.Unlock()

		c.JSON(http.StatusOK, gin.H{
			"err":   "0",
			"nonce": base64.RawURLEncoding.EncodeToString(nonce),
		})
	}
}

func PassChallengeV1(deps *api.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		bankID := c.Param("bankId")

		memory.Mutex.Lock()

		bank, ok := memory.Banks[bankID]

		challenge, ok := memory.Challenges[bankID]

		memory.Mutex.Unlock()

		if !ok {
			c.JSON(http.StatusNotFound, gin.H{
				"err": "NO_BANK",
				"errmsg": "we don't have information about this bank; " +
					"first send public key via POST /api/bank/{bankID}/pk",
			})
			return
		}

		if !ok || time.Now().After(challenge.ExpiresAt) {
			c.JSON(http.StatusBadRequest, gin.H{
				"err":    "EXPIRED_CHALLENGE",
				"errmsg": "this challenge has expired",
			})
			return
		}

		var req struct {
			SignatureB64 string `json:"s"`
		}

		if c.BindJSON(&req) != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"err":    "CANT_BIND_JSON",
				"errmsg": "could not bind request body",
			})
			return
		}

		signature, err := base64.RawURLEncoding.DecodeString(req.SignatureB64)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"err":    "CANT_BIND_JSON",
				"errmsg": "could not bind request body",
			})
			return
		}

		publicKey := slhdsa.PublicKey{ID: slhdsa.SHAKE_256s}
		err = publicKey.UnmarshalBinary(bank.RootPublicKeyBytes)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"err":    "INTERNAL_ERROR",
				"errmsg": "could not decode public key",
			})
			return
		}

		domain := "BESHENCE-BANK-GATEWAY-PASS-CHALLENGE-V1"

		message := make([]byte, 0, len(domain)+len(challenge.Nonce))

		message = append(
			message,
			[]byte(domain)...,
		)

		message = append(
			message,
			challenge.Nonce...,
		)

		valid := slhdsa.Verify(
			&publicKey,
			slhdsa.NewMessage(message),
			signature,
			nil,
		)

		if !valid {
			c.JSON(http.StatusBadRequest, gin.H{
				"err":    "CANT_VERIFY_SIGNATURE",
				"errmsg": "could not verify signature",
			})
			return
		}

		jwtToken, err := deps.JWTManager.GenerateToken(bankID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"err":    "INTERNAL_ERROR",
				"errmsg": "internal error when generating JWT token",
			})
			return
		}

		memory.Mutex.Lock()

		delete(memory.Challenges, bankID)

		memory.Mutex.Unlock()

		c.JSON(http.StatusOK, gin.H{
			"err":   "0",
			"token": jwtToken,
		})
	}
}
