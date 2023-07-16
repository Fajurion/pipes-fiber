package pipesfroutes

import (
	"chat-node/util/requests"

	"github.com/Fajurion/pipes"
	"github.com/Fajurion/pipes/connection"
	"github.com/Fajurion/pipes/receive"
	"github.com/Fajurion/pipesfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func adoptionRouter(router fiber.Router) {
	router.Use("/", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {

			// Check if the request has a token
			token := c.Get("Sec-WebSocket-Protocol")

			// Adopt node
			node, err := receive.ReceiveWSAdoption(token)
			if err != nil {
				return requests.InvalidRequest(c)
			}

			// Set the token as a local variable
			c.Locals("ws", true)
			c.Locals("node", node)
			return c.Next()
		}

		return c.SendStatus(fiber.StatusUpgradeRequired)
	})

	router.Get("/", websocket.New(adoptionWs))
}

func adoptionWs(conn *websocket.Conn) {
	node := conn.Locals("node").(pipes.Node)

	defer func() {

		// Disconnect node
		connection.RemoveWS(node.ID)
		pipesfiber.CurrentConfig.NodeDisconnectHandler(node)
		// TODO: integration.ReportOffline(node)
		conn.Close()
	}()

	for {
		// Read message as text
		mtype, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}

		if mtype == websocket.TextMessage {

			// Pass message to pipes
			receive.ReceiveWS(msg)
		}
	}

}
