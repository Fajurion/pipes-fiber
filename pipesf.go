package pipesfiber

import (
	pipesfcache "github.com/Fajurion/pipesfiber/caching"
	pipesfutil "github.com/Fajurion/pipesfiber/util"
	"github.com/gofiber/fiber/v2"
)

func Setup(config pipesfutil.Config) {
	pipesfutil.CurrentConfig = config
	pipesfcache.SetupConnectionsCache(config.ExpectedConnections)
	pipesfcache.SetupTokenCache(config.ExpectedConnections)
}

func SetupRoutes(router fiber.Router) {
	router.Route("/gateway", gatewayRouter) // gatewayRouter is a function in gateway.go

	router.Post("/adoption/socketless", socketless) // socketless is a function in socketless.go
	router.Route("/adoption", adoptionRouter)       // adoption is a function in adoption.go
}
