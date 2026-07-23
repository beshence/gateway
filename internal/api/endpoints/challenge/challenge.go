package challenge

import (
	"crypto/hmac"
	"crypto/mlkem"
	"encoding/base64"
	"gateway/internal/api"
	"gateway/internal/memory"
	"gateway/internal/misc"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func GetChallengeV1() gin.HandlerFunc {
	return func(c *gin.Context) {
		bankID := c.Param("bankId")

		memory.Mutex.Lock()

		bank, ok := memory.Banks[bankID]

		memory.Mutex.Unlock()

		if !ok {
			c.JSON(http.StatusNotFound, gin.H{
				"err":    "NO_BANK",
				"errmsg": "we don't have information about this bank; first send encapsulation key via POST /api/bank/{bankID}/ek",
			})
			return
		}

		memory.Mutex.Lock()

		challenge, ok := memory.Challenges[bankID]

		memory.Mutex.Unlock()

		if ok {
			if challenge.ExpiresAt.After(time.Now()) {
				c.JSON(http.StatusOK, gin.H{
					"err":        "0",
					"ciphertext": base64.RawURLEncoding.EncodeToString(challenge.Ciphertext),
				})
				return
			}
		}

		ekBytes := bank.EK

		ek, err := mlkem.NewEncapsulationKey1024(ekBytes)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"err":    "CANT_READ_EK",
				"errmsg": "internal error when reading encapsulation key",
			})
		}

		sharedSecret, ciphertext := ek.Encapsulate()

		memory.Mutex.Lock()

		memory.Banks[bankID] = memory.Bank{
			EK: ek.Bytes(),
		}

		memory.Challenges[bankID] = memory.Challenge{
			Ciphertext: ciphertext,
			Secret:     sharedSecret,
			ExpiresAt:  time.Now().Add(time.Minute),
		}

		memory.Mutex.Unlock()

		c.JSON(http.StatusOK, gin.H{
			"err":        "0",
			"ciphertext": base64.RawURLEncoding.EncodeToString(ciphertext),
		})
		return
	}
}

func PassChallengeV1(deps *api.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		bankID := c.Param("bankId")

		var req struct {
			Proof string `json:"proof"`
		}

		if c.BindJSON(&req) != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"err":    "CANT_BIND_JSON",
				"errmsg": "could not bind request body",
			})
			return
		}

		proof, err := base64.RawURLEncoding.DecodeString(req.Proof)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"err":    "CANT_DECODE_PROOF",
				"errmsg": "could not decode proof",
			})
			return
		}

		memory.Mutex.Lock()

		challenge, ok := memory.Challenges[bankID]

		memory.Mutex.Unlock()

		if !ok || time.Now().After(challenge.ExpiresAt) {
			c.JSON(http.StatusBadRequest, gin.H{
				"err":    "EXPIRED_CHALLENGE",
				"errmsg": "this challenge has expired",
			})
			return
		}

		expectedProof := misc.MakeProof(
			challenge.Secret,
			challenge.Ciphertext,
		)

		if !hmac.Equal(expectedProof, proof) {
			c.JSON(http.StatusBadRequest, gin.H{
				"err":    "WRONG_PROOF",
				"errmsg": "your proof and expected proof are different",
			})
			return
		}

		jwtToken, err := deps.JWTManager.GenerateToken(bankID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"err":    "CANT_ISSUE_JWT",
				"errmsg": "internal error when generating JWT token",
			})
			return
		}

		memory.Mutex.Lock()

		delete(
			memory.Challenges,
			bankID,
		)

		memory.Mutex.Unlock()

		c.JSON(http.StatusOK, gin.H{
			"err":   "0",
			"token": jwtToken,
		})
	}
}
