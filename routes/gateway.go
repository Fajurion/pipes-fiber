package pipesfroutes

import (
	"errors"
	"fmt"
	"log"
	"strings"
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
			protocolSeperated := c.Get("Sec-WebSocket-Protocol")
			protocols := strings.Split(protocolSeperated, ", ")
			token := protocols[0]

			// Get attachments from the connection (passed to the node)
			attachments := ""
			if len(protocols) > 1 {
				attachments = protocols[1]
			}

			if len(token) == 0 {
				return c.SendStatus(fiber.StatusUnauthorized)
			}

			log.Println(c.GetReqHeaders())

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
			c.Locals("attached", attachments)
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
		adapter.RemoveWS(tk.UserID)
		pipesfiber.CurrentConfig.ClientDisconnectHandler(client)

		// Remove the connection from the cache
		pipesfiber.Remove(tk.UserID, tk.Session)
	}()

	if pipesfiber.CurrentConfig.ClientConnectHandler(client, conn.Locals("attached").(string)) {
		return
	}

	// Add adapter for pipes
	adapter.AdaptWS(adapter.Adapter{
		ID: tk.UserID,
		Receive: func(c *adapter.Context) error {

			// Get the client
			client, valid := pipesfiber.Get(tk.UserID, tk.Session)
			if !valid {
				pipesfiber.ReportGeneralError("couldn't get client", fmt.Errorf("%s (%s)", tk.UserID, tk.Session))
				return errors.New("couldn't get client")
			}

			// Send message encoded with client encoding middleware
			msg, err := pipesfiber.CurrentConfig.ClientEncodingMiddleware(client, c.Message)
			if err != nil {
				pipesfiber.ReportClientError(client, "couldn't encode received message", err)
				return err
			}

			log.Println("sending " + c.Event.Name)

			return conn.WriteMessage(websocket.BinaryMessage, msg)
		},
	})

	if pipesfiber.CurrentConfig.ClientEnterNetworkHandler(client, conn.Locals("attached").(string)) {
		return
	}

	for {

		// Read message as text
		_, msg, err := conn.ReadMessage()
		if err != nil {

			// Get the client for error reporting purposes
			client, valid := pipesfiber.Get(tk.UserID, tk.Session)
			if !valid {
				pipesfiber.ReportGeneralError("couldn't get client", fmt.Errorf("%s (%s)", tk.UserID, tk.Session))
				return
			}

			pipesfiber.ReportClientError(client, "couldn't read message", err)
			break
		}

		// Get the client
		client, valid := pipesfiber.Get(tk.UserID, tk.Session)
		if !valid {
			pipesfiber.ReportGeneralError("couldn't get client", fmt.Errorf("%s (%s)", tk.UserID, tk.Session))
			return
		}

		// Unmarshal the action
		message, err := pipesfiber.CurrentConfig.DecodingMiddleware(client, msg)
		if err != nil {
			pipesfiber.ReportClientError(client, "couldn't decode message", err)
			return
		}

		if client.IsExpired() {
			return
		}

		// Handle the action
		if !wshandler.Handle(wshandler.Message{
			Client: client,
			Data:   message.Data,
			Action: message.Action,
		}) {
			pipesfiber.ReportClientError(client, "couldn't handle action", errors.New(message.Action))
			return
		}
	}
}
