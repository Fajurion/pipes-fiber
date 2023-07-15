package pipesfcache

import (
	"time"

	"github.com/dgraph-io/ristretto"
)

type ConnectionToken struct {
	UserID   string
	Session  string
	Username string
	Tag      string
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

func AddToken(token string, id string, session string, username string, tag string) {

	tokenCache.SetWithTTL(token, ConnectionToken{
		UserID:   id,
		Session:  session,
		Username: username,
		Tag:      tag,
	}, 1, TokenTTL)
}
