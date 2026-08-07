package urls

import (
	"gateway/internal/auth"
	"gateway/internal/memory"
	"net/http"

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

func PostApiUrlsV1() gin.HandlerFunc {
	return func(c *gin.Context) {
		bankID := c.Param("bankId")

		claimsBankID, ok := auth.GetCurrentBank(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"err":    "UNAUTHORIZED",
				"errmsg": "unauthorized",
			})
			return
		}

		if bankID != claimsBankID {
			c.JSON(http.StatusUnauthorized, gin.H{
				"err":    "UNAUTHORIZED",
				"errmsg": "unauthorized",
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

		var req struct {
			ApiUrls []string `json:"api_urls"`
		}

		if c.BindJSON(&req) != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"err":    "CANT_BIND_JSON",
				"errmsg": "could not bind request body",
			})
			return
		}

		var bankApiUrls memory.BankApiUrls

		bankApiUrls.Urls = req.ApiUrls

		memory.Mutex.Lock()

		memory.BanksApiUrls[bankID] = bankApiUrls

		memory.Mutex.Unlock()

		c.JSON(http.StatusCreated, gin.H{
			"err": "0",
		})
	}
}
