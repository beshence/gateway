package ws

import (
	"gateway/internal/api"
	"gateway/internal/auth"
	"gateway/internal/signal"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/websocket"
)

func WSV1(deps *api.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		bankID := c.Param("bankId")
		sessionID := c.Query("session_id")

		if bankID == "" {
			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"err": "MISSING_BANK_ID",
				},
			)
			return
		}

		role := signal.PeerRoleClient

		if c.GetHeader("Authorization") != "" {
			auth.CheckAuth(deps.JWTManager)(c)

			if c.IsAborted() {
				return
			}

			tokenBankID, ok := auth.GetCurrentBank(c)

			if !ok || tokenBankID != bankID {
				c.JSON(
					http.StatusUnauthorized,
					gin.H{
						"err": "UNAUTHORIZED",
					},
				)
				return
			}

			if sessionID != "" {
				c.JSON(
					http.StatusForbidden,
					gin.H{
						"err": "FORBIDDEN",
					},
				)
				return
			}
			role = signal.PeerRoleBank
		} else {
			if sessionID == "" {
				c.JSON(
					http.StatusBadRequest,
					gin.H{
						"err": "MISSING_SESSION_ID",
					},
				)
				return
			}
		}

		handler :=
			websocket.Handler(
				func(conn *websocket.Conn) {
					peer :=
						&signal.Peer{
							BankID:    bankID,
							Role:      role,
							SessionID: sessionID,
							Conn:      conn,
						}

					deps.SignalManager.Add(peer)

					defer func() {
						deps.SignalManager.Remove(peer)
						err := conn.Close()
						if err != nil {
							return
						}
					}()

					for {
						var message signal.Message
						err := websocket.JSON.Receive(conn, &message)
						if err != nil {
							break
						}
						deps.SignalManager.Forward(peer, message)
					}
				},
			)

		handler.ServeHTTP(
			c.Writer,
			c.Request,
		)
	}
}
