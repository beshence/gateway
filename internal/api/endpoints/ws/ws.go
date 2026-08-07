package ws

import (
	"context"
	"gateway/internal/api"
	"gateway/internal/auth"
	"gateway/internal/signal"
	"net/http"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/gin-gonic/gin"
)

func WSV1(deps *api.Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		bankID := c.Param("bankId")
		roleRaw := c.Query("role")
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

		if roleRaw == "" {
			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"err": "MISSING_ROLE",
				},
			)
			return
		}

		if roleRaw != "client" && roleRaw != "bank" {
			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"err": "INVALID_ROLE",
				},
			)
			return
		}

		var role signal.PeerRole

		if roleRaw == "bank" {
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
						"err": "INVALID_SESSION_ID",
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

			role = signal.PeerRoleClient
		}

		conn, err := websocket.Accept(
			c.Writer,
			c.Request,
			&websocket.AcceptOptions{
				InsecureSkipVerify: true,
			},
		)

		if err != nil {
			return
		}

		defer conn.Close(
			websocket.StatusNormalClosure,
			"",
		)

		peer := &signal.Peer{
			BankID:    bankID,
			Role:      role,
			SessionID: sessionID,
			Conn:      conn,
		}

		deps.SignalManager.Add(peer)

		defer func() {
			deps.SignalManager.Remove(peer)
		}()

		ctx := context.Background()

		for {
			var message signal.Message

			err := wsjson.Read(
				ctx,
				conn,
				&message,
			)

			if err != nil {
				break
			}

			deps.SignalManager.Forward(peer, message)
		}
	}
}
