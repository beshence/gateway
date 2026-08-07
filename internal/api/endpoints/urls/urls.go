package urls

import (
	"encoding/base64"
	"encoding/json"
	"gateway/internal/memory"
	"net/http"
	"time"

	"github.com/cloudflare/circl/sign/slhdsa"
	"github.com/gin-gonic/gin"
)

func GetApiUrlsV1() gin.HandlerFunc {
	return func(c *gin.Context) {
		bankID := c.Param("bankId")

		memory.Mutex.Lock()

		data, ok := memory.BanksApiUrls[bankID]

		memory.Mutex.Unlock()

		if !ok {
			c.JSON(http.StatusNotFound, gin.H{
				"err":    "NO_BANK",
				"errmsg": "we don't have information about this bank",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"err":  "0",
			"urls": data.Urls,
		})
	}
}

func GetApiUrlsPayloadSignatureV1() gin.HandlerFunc {
	return func(c *gin.Context) {
		bankID := c.Param("bankId")

		memory.Mutex.Lock()

		data, ok := memory.BanksApiUrls[bankID]

		memory.Mutex.Unlock()

		if !ok {
			c.JSON(http.StatusNotFound, gin.H{
				"err":    "NO_BANK",
				"errmsg": "we don't have information about this bank",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"err": "0",
			"p":   data.PayloadB64,
			"s":   data.SignatureB64,
		})
	}
}

type postApiUrlsV1Request struct {
	PayloadB64   string `json:"p" binding:"required"`
	SignatureB64 string `json:"s" binding:"required"`
}

func PostApiUrlsV1() gin.HandlerFunc {
	return func(c *gin.Context) {
		bankID := c.Param("bankId")

		memory.Mutex.Lock()

		bank, ok := memory.Banks[bankID]

		memory.Mutex.Unlock()

		if !ok {
			c.JSON(http.StatusNotFound, gin.H{
				"err": "NO_BANK",
				"errmsg": "we don't have information about this bank; " +
					"first send public key via POST /api/bank/{bankID}/pk",
			})
			return
		}

		var req postApiUrlsV1Request

		if c.BindJSON(&req) != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"err":    "CANT_BIND_JSON",
				"errmsg": "could not bind request body",
			})
			return
		}

		payloadBytes, err := base64.RawURLEncoding.DecodeString(req.PayloadB64)

		if err != nil {
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

		type Payload struct {
			Urls      []string `json:"urls"`
			ExpiresAt int64    `json:"expires_at"`
		}

		var payload Payload
		err = json.Unmarshal(payloadBytes, &payload)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"err":    "CANT_DECODE_PAYLOAD",
				"errmsg": "could not decode payload",
			})
			return
		}

		if payload.ExpiresAt < time.Now().UnixMilli() {
			c.JSON(http.StatusBadRequest, gin.H{
				"err":    "EXPIRED_PAYLOAD",
				"errmsg": "payload has expired",
			})
			return
		}

		publicKey := slhdsa.PublicKey{ID: slhdsa.SHAKE_256s}
		err = publicKey.UnmarshalBinary(bank.PublicKeyBytes)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"err":    "INTERNAL_ERROR",
				"errmsg": "could not decode public key",
			})
			return
		}

		domain := "BESHENCE-BANK-GATEWAY-POST-URLS-V1"

		message := make([]byte, 0, len(domain)+len(payloadBytes))

		message = append(
			message,
			[]byte(domain)...,
		)

		message = append(
			message,
			payloadBytes...,
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

		if !memory.GetLimiter(bankID).Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"err":    "RATE_LIMIT",
				"errmsg": "rate limited",
			})
			return
		}

		var bankApiUrls memory.BankApiUrls
		bankApiUrls.Urls = payload.Urls
		bankApiUrls.PayloadB64 = req.PayloadB64
		bankApiUrls.SignatureB64 = req.SignatureB64

		memory.Mutex.Lock()

		memory.BanksApiUrls[bankID] = bankApiUrls

		memory.Mutex.Unlock()

		c.JSON(http.StatusCreated, gin.H{
			"err": "0",
		})
	}
}
