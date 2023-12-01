package pipesfiber

import (
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/gofiber/websocket/v2"
)

type ConnectionToken struct {
	UserID  string
	Session string
	Data    interface{}
}

func (tk ConnectionToken) ToClient(conn *websocket.Conn, end time.Time) Client {
	return Client{
		Conn:    conn,
		ID:      tk.UserID,
		Session: tk.Session,
		End:     end,
		Data:    tk.Data,
		Mutex:   &sync.Mutex{},
	}
}

// ! Cost 1 for all caches
var tokenCache *ristretto.Cache

// * Time to live for tokens
const TokenTTL = time.Hour * 1

func SetupTokenCache(expected int64) {

	var err error
	tokenCache, err = ristretto.NewCache(&ristretto.Config{
		NumCounters: expected,               // expecting to store 10k connections
		MaxCost:     expected - expected/10, // maximum items in the cache (with cost 1 on each set)
		BufferItems: 64,                     // Some random number, check docs
	})

	if err != nil {
		panic(err)
	}

}

func CheckToken(token string) (ConnectionToken, bool) {

	tk, ok := tokenCache.Get(token)
	if !ok {
		return ConnectionToken{}, false
	}

	return tk.(ConnectionToken), ok
}

func RemoveToken(token string) {
	tokenCache.Del(token)
}

func AddToken(tk string, token ConnectionToken) {

	tokenCache.SetWithTTL(tk, token, 1, TokenTTL)
}
