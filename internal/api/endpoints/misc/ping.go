package misc

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func PingV1() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"err":  "0",
			"ping": "beshence-gateway-pong!",
		})
	}
}
