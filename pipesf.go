package pipesfiber

import (
	"time"

	"github.com/Fajurion/pipes"
)

var CurrentConfig = Config{
	ExpectedConnections: 1000,
}

// ! If the functions aren't implemented pipesfiber will panic
type Config struct {
	ExpectedConnections int64
	SessionDuration     time.Duration // How long a session should stay alive

	// Node handlers
	NodeDisconnectHandler func(node pipes.Node)

	// Client handlers
	ClientDisconnectHandler   func(client *Client)
	ClientConnectHandler      func(client *Client) bool // Returns if the client should be disconnected (true = disconnect)
	ClientEnterNetworkHandler func(client *Client) bool // Returns if the client should be disconnected (true = disconnect)
}

func Setup(config Config) {
	CurrentConfig = config
	SetupConnectionsCache(config.ExpectedConnections)
	SetupTokenCache(config.ExpectedConnections)
}
