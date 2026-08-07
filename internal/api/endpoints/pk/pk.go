package pk

import (
	"encoding/base64"
	"gateway/internal/memory"
	"gateway/internal/misc"
	"net/http"

	"github.com/cloudflare/circl/sign/slhdsa"
	"github.com/gin-gonic/gin"
)

func GetPublicKeyV1() gin.HandlerFunc {
	return func(c *gin.Context) {
		bankID := c.Param("bankId")

		memory.Mutex.Lock()

		bank, ok := memory.Banks[bankID]

		memory.Mutex.Unlock()

		if !ok {
			c.JSON(http.StatusNotFound, gin.H{
				"err":    "NO_BANK",
				"errmsg": "we don't have public key of this bank",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"err": "0",
			"pk":  base64.RawURLEncoding.EncodeToString(bank.PublicKeyBytes),
		})
	}
}

type postPublicKeyV1Request struct {
	PublicKeyEncoded string `json:"pk" binding:"required"`
}

func PostPublicKeyV1() gin.HandlerFunc {
	return func(c *gin.Context) {
		bankID := c.Param("bankId")

		var req postPublicKeyV1Request

		if c.BindJSON(&req) != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"err":    "CANT_BIND_JSON",
				"errmsg": "could not bind request body",
			})
			return
		}

		publicKeyBytes, err := base64.RawURLEncoding.DecodeString(req.PublicKeyEncoded)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"err":    "CANT_DECODE_PK",
				"errmsg": "could not decode public key",
			})
			return
		}

		publicKey := slhdsa.PublicKey{ID: slhdsa.SHAKE_256s}

		err = publicKey.UnmarshalBinary(publicKeyBytes)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"err":    "CANT_DECODE_PK",
				"errmsg": "could not decode public key",
			})
			return
		}

		generatedBankID, _ := misc.GetBankID(publicKey)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"err":    "CANT_DECODE_PK",
				"errmsg": "could not decode public key",
			})
			return
		}

		if bankID != generatedBankID {
			c.JSON(http.StatusBadRequest, gin.H{
				"err":    "CANT_DECODE_PK",
				"errmsg": "could not decode public key",
			})
			return
		}

		memory.Mutex.Lock()

		memory.Banks[bankID] = memory.Bank{
			PublicKeyBytes: publicKeyBytes,
		}

		memory.Mutex.Unlock()

		c.JSON(http.StatusCreated, gin.H{
			"err": "0",
		})
	}
}
