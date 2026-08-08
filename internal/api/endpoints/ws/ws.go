package ws

import (
	"context"
	"gateway/internal/api"
	"gateway/internal/auth"
	"gateway/internal/websocket"
	"net/http"

	libws "github.com/coder/websocket"
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

		var role websocket.PeerRole

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

			role = websocket.PeerRoleBank

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

			role = websocket.PeerRoleClient
		}

		conn, err := libws.Accept(
			c.Writer,
			c.Request,
			&libws.AcceptOptions{
				InsecureSkipVerify: true,
			},
		)

		if err != nil {
			return
		}

		defer conn.Close(
			libws.StatusNormalClosure,
			"",
		)

		peer := &websocket.Peer{
			BankID:    bankID,
			Role:      role,
			SessionID: sessionID,
			Conn:      conn,
		}

		deps.WebSocketManager.Add(peer)

		defer func() {
			deps.WebSocketManager.Remove(peer)
		}()

		ctx := context.Background()

		for {
			var message websocket.Message

			err := wsjson.Read(
				ctx,
				conn,
				&message,
			)

			if err != nil {
				break
			}

			deps.WebSocketManager.Forward(peer, message)
		}
	}
}
