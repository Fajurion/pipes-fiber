package pipesfroutes

import "github.com/gofiber/fiber/v2"

func SetupRoutes(router fiber.Router) {
	router.Route("/gateway", gatewayRouter) // gatewayRouter is a function in gateway.go

	router.Post("/adoption/socketless", socketless) // socketless is a function in socketless.go
	router.Route("/adoption", adoptionRouter)       // adoption is a function in adoption.go
}
