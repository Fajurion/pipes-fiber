package pipesfroutes

import (
	"time"

	"github.com/Fajurion/pipes/adapter"
	"github.com/Fajurion/pipesfiber"
	"github.com/Fajurion/pipesfiber/wshandler"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func gatewayRouter(router fiber.Router) {

	// Inject a middleware to check if the request is a websocket upgrade request
	router.Use("/", func(c *fiber.Ctx) error {

		if websocket.IsWebSocketUpgrade(c) {

			// Check if the request has a token
			token := c.Get("Sec-WebSocket-Protocol")

			if len(token) == 0 {
				return c.SendStatus(fiber.StatusUnauthorized)
			}

			// Check if the token is valid
			tk, ok := pipesfiber.CheckToken(token)
			if !ok {
				return c.SendStatus(fiber.StatusBadRequest)
			}

			if pipesfiber.ExistsConnection(tk.UserID, tk.Session) {
				return c.SendStatus(fiber.StatusConflict)
			}

			pipesfiber.RemoveToken(token)

			// Set the token as a local variable
			c.Locals("ws", true)
			c.Locals("tk", tk)
			return c.Next()
		}

		return c.SendStatus(fiber.StatusUpgradeRequired)
	})

	router.Get("/", websocket.New(ws))
}

func ws(conn *websocket.Conn) {
	tk := conn.Locals("tk").(pipesfiber.ConnectionToken)

	client := pipesfiber.AddClient(tk.ToClient(conn, time.Now().Add(pipesfiber.CurrentConfig.SessionDuration)))
	defer func() {

		// Send callback to app
		client, valid := pipesfiber.Get(tk.UserID, tk.Session)
		if !valid {
			return
		}
		pipesfiber.CurrentConfig.ClientDisconnectHandler(client)

		// Remove the connection from the cache
		pipesfiber.Remove(tk.UserID, tk.Session)
	}()

	if pipesfiber.CurrentConfig.ClientConnectHandler(client) {
		return
	}

	// Add adapter for pipes
	adapter.AdaptWS(adapter.Adapter{
		ID: tk.UserID,
		Receive: func(c *adapter.Context) error {
			return conn.WriteMessage(websocket.TextMessage, c.Message)
		},
	})

	if pipesfiber.CurrentConfig.ClientEnterNetworkHandler(client) {
		return
	}

	for {
		// Read message as text
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// Get the client
		client, valid := pipesfiber.Get(tk.UserID, tk.Session)
		if !valid {
			return
		}

		// Unmarshal the event
		message, err := pipesfiber.CurrentConfig.DecodingMiddleware(client, msg)
		if err != nil {
			return
		}

		if client.IsExpired() {
			return
		}

		// Handle the event
		if !wshandler.Handle(wshandler.Message{
			Client: client,
			Data:   message.Data,
			Action: message.Action,
		}) {
			return
		}
	}
}
