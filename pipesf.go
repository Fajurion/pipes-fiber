package pipesfiber

import (
	"fmt"
	"time"

	"github.com/Fajurion/pipes"
	"github.com/bytedance/sonic"
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

	// Codec middleware
	ClientEncodingMiddleware func(client *Client, message []byte) ([]byte, error)
	DecodingMiddleware       func(client *Client, message []byte) (Message, error)

	// Error handler
	ErrorHandler func(err error)
}

// Message received from the client
type Message struct {
	Action string                 `json:"action"`
	Data   map[string]interface{} `json:"data"`
}

// Default pipes-fiber encoding middleware (using JSON)
func DefaultClientEncodingMiddleware(client *Client, message []byte) ([]byte, error) {
	return message, nil
}

// Default pipes-fiber decoding middleware (using JSON)
func DefaultDecodingMiddleware(client *Client, bytes []byte) (Message, error) {
	var message Message
	if err := sonic.Unmarshal(bytes, &message); err != nil {
		return Message{}, err
	}
	return message, nil
}

func Setup(config Config) {
	CurrentConfig = config
	SetupConnectionsCache(config.ExpectedConnections)
	SetupTokenCache(config.ExpectedConnections)
}

func ReportGeneralError(context string, err error) {
	if CurrentConfig.ErrorHandler == nil {
		return
	}

	CurrentConfig.ErrorHandler(fmt.Errorf("general: %s: %s", context, err.Error()))
}

func ReportClientError(client *Client, context string, err error) {
	if CurrentConfig.ErrorHandler == nil {
		return
	}

	CurrentConfig.ErrorHandler(fmt.Errorf("client %s: %s: %s", client.ID, context, err.Error()))
}
