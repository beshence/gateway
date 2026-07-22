package challenge

import (
	"crypto/hmac"
	"crypto/mlkem"
	"encoding/base64"
	"gateway/internal/api"
	"gateway/internal/memory"
	"gateway/internal/misc"
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
			c.Status(404)
			return
		}

		memory.Mutex.Lock()

		challenge, ok := memory.Challenges[bankID]

		memory.Mutex.Unlock()

		if ok {
			if challenge.ExpiresAt.After(time.Now()) {
				c.JSON(200, gin.H{
					"err":        "0",
					"ciphertext": base64.RawURLEncoding.EncodeToString(challenge.Ciphertext),
				})
				return
			}
		}

		ekBytes := bank.EK

		ek, err := mlkem.NewEncapsulationKey1024(ekBytes)

		if err != nil {
			c.Status(500)
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

		c.JSON(200, gin.H{
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
			c.Status(400)
			return
		}

		proof, err := base64.RawURLEncoding.DecodeString(req.Proof)

		if err != nil {
			c.Status(400)
			return
		}

		memory.Mutex.Lock()

		challenge, ok := memory.Challenges[bankID]

		memory.Mutex.Unlock()

		if !ok || time.Now().After(challenge.ExpiresAt) {
			c.Status(401)
			return
		}

		expectedProof := misc.MakeProof(
			challenge.Secret,
			challenge.Ciphertext,
		)

		if !hmac.Equal(expectedProof, proof) {
			c.Status(403)
			return
		}

		jwtToken, err := deps.JWTManager.GenerateToken(bankID)

		if err != nil {
			c.Status(500)
			return
		}

		memory.Mutex.Lock()

		delete(
			memory.Challenges,
			bankID,
		)

		memory.Mutex.Unlock()

		c.JSON(200, gin.H{
			"err":   "0",
			"token": jwtToken,
		})
	}
}
