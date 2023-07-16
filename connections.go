package pipesfiber

import (
	"chat-node/database"
	"chat-node/database/fetching"
	"chat-node/util"
	"time"

	"github.com/Fajurion/pipes"
	pipesfutil "github.com/Fajurion/pipesfiber/util"
	"github.com/bytedance/sonic"
	"github.com/dgraph-io/ristretto"
	"github.com/gofiber/websocket/v2"
)

type Client struct {
	Conn    *websocket.Conn
	ID      string
	Session string
	End     time.Time
	Data    interface{}
}

func (c *Client) SendEvent(event pipes.Event) {

	event.Sender = c.ID
	msg, err := sonic.Marshal(event)
	if err != nil {
		return
	}

	SendMessage(c.Conn, msg)
}

func (c *Client) IsExpired() bool {
	return c.End.Before(time.Now())
}

// ! Cost 1 for all caches
// ID:Session -> Client
var connectionsCache *ristretto.Cache

// ID -> Session list
var sessionsCache *ristretto.Cache

func SetupConnectionsCache(expected int64) {

	var err error
	connectionsCache, err = ristretto.NewCache(&ristretto.Config{
		NumCounters: expected,               // expecting to store 10k connections
		MaxCost:     expected - expected/10, // maximum items in the cache (with cost 1 on each set)
		BufferItems: 64,                     // Some random number, check docs
	})

	if err != nil {
		panic(err)
	}

	sessionsCache, err = ristretto.NewCache(&ristretto.Config{
		NumCounters: expected,               // expecting to store 10k connections
		MaxCost:     expected - expected/10, // maximum items in the cache (with cost 1 on each set)
		BufferItems: 64,                     // Some random number, check docs
	})

	if err != nil {
		panic(err)
	}

}

func getKey(id string, session string) string {
	return id + ":" + session
}

func AddClient(client Client) *Client {

	_, add := connectionsCache.Get(getKey(client.ID, client.Session))
	connectionsCache.Set(getKey(client.ID, client.Session), client, 1)

	if add {
		addSession(client.ID, client.Session)
	}

	return &client
}

func GetSessions(id string) []string {
	sessions, valid := sessionsCache.Get(id)
	if valid {
		return sessions.([]string)
	}

	return []string{}
}

func addSession(id string, session string) {

	sessions, valid := sessionsCache.Get(id)
	if valid {
		sessionsCache.Set(id, append(sessions.([]string), session), 1)
	} else {
		sessionsCache.Set(id, []string{session}, 1)
	}
}

func removeSession(id string, session string) {

	sessions, valid := sessionsCache.Get(id)
	if valid {

		if len(sessions.([]string)) == 1 {
			sessionsCache.Del(id)
			return
		}

		sessionsCache.Set(id, pipesfutil.RemoveString(sessions.([]string), session), 1)
	}
}

func Remove(id string, session string) {

	database.DBConn.Model(&fetching.Session{}).Where("id = ?", session).Update("last_fetch", time.Now().UnixMilli())
	util.PostRequest("/node/disconnect", map[string]interface{}{
		"node":    util.NODE_ID,
		"token":   util.NODE_TOKEN,
		"session": session,
	})

	connectionsCache.Del(getKey(id, session))
	removeSession(id, session)
}

func Send(id string, msg []byte) {
	sessions, ok := sessionsCache.Get(id)

	if !ok {
		return
	}

	for _, session := range sessions.([]string) {
		client, valid := Get(id, session)
		if !valid {
			continue
		}

		SendMessage(client.Conn, msg)
	}
}

func SendSession(id string, session string, msg []byte) bool {
	client, valid := Get(id, session)
	if !valid {
		return false
	}

	SendMessage(client.Conn, msg)
	return true
}

func SendMessage(conn *websocket.Conn, msg []byte) {
	conn.WriteMessage(websocket.TextMessage, msg)
}

func ExistsConnection(id string, session string) bool {
	_, ok := connectionsCache.Get(getKey(id, session))
	if !ok {
		return false
	}

	return ok
}

func Get(id string, session string) (*Client, bool) {
	client, valid := connectionsCache.Get(getKey(id, session))
	if !valid {
		return &Client{}, false
	}

	cl := client.(Client)
	return &cl, true
}

func GetConnections(id string) int {
	clients, ok := sessionsCache.Get(id)
	if !ok {
		return 0
	}

	return len(clients.([]string))
}
