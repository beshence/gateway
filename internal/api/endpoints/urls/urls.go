package urls

import (
	"gateway/internal/auth"
	"gateway/internal/memory"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func GetAPIURLsV1() gin.HandlerFunc {
	return func(c *gin.Context) {
		bankID := c.Param("bankId")

		memory.Mutex.Lock()

		data, ok := memory.APIURLss[bankID]

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
			"urls": data.ApiUrls,
		})
	}
}

func SetAPIURLsV1() gin.HandlerFunc {
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

		var apiUrls memory.APIURLs

		apiUrls.ApiUrls = req.ApiUrls
		apiUrls.UpdatedAt = time.Now()

		memory.Mutex.Lock()

		memory.APIURLss[bankID] = apiUrls

		memory.Mutex.Unlock()

		c.JSON(http.StatusCreated, gin.H{
			"err": "0",
		})
	}
}
