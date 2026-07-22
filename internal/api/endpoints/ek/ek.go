package ek

import (
	"crypto/mlkem"
	"encoding/base64"
	"gateway/internal/memory"
	"gateway/internal/misc"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetEKV1() gin.HandlerFunc {
	return func(c *gin.Context) {
		bankID := c.Param("bankId")

		memory.Mutex.Lock()

		bank, ok := memory.Banks[bankID]

		memory.Mutex.Unlock()

		if !ok {
			c.Status(404)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"err": "0",
			"ek":  base64.RawURLEncoding.EncodeToString(bank.EK),
		})
	}
}

type postEKV1Request struct {
	EK string `json:"ek" binding:"required"`
}

func PostEKV1() gin.HandlerFunc {
	return func(c *gin.Context) {
		bankID := c.Param("bankId")

		var req postEKV1Request

		if c.BindJSON(&req) != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		ekBytes, err := base64.RawURLEncoding.DecodeString(req.EK)

		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		ek, err := mlkem.NewEncapsulationKey1024(ekBytes)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		generatedBankID := misc.GetBankID(ek)

		if bankID != generatedBankID {
			c.Status(401)
			return
		}

		memory.Mutex.Lock()

		memory.Banks[bankID] = memory.Bank{
			EK: ek.Bytes(),
		}

		memory.Mutex.Unlock()

		c.JSON(201, gin.H{
			"err": "0",
		})
	}
}
