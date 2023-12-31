package pipesfroutes

import (
	"github.com/Fajurion/pipes"
	"github.com/Fajurion/pipes/receive"
	"github.com/Fajurion/pipes/send"
	"github.com/gofiber/fiber/v2"
)

type socketlessEvent struct {
	Token   string        `json:"token"`
	Message pipes.Message `json:"message"`
}

func socketless(c *fiber.Ctx) error {

	// Parse request
	var event socketlessEvent
	if err := c.BodyParser(&event); err != nil {
		return err
	}

	// Check token
	if event.Token != pipes.CurrentNode.Token {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	receive.HandleMessage(send.ProtocolWS, event.Message)

	return c.JSON(fiber.Map{
		"success": true,
	})
}
