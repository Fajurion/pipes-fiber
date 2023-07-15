package pipesfiber

import pipesfcache "github.com/Fajurion/pipesfiber/caching"

func Setup(expectedConnections int64) {
	pipesfcache.SetupConnectionsCache(expectedConnections)
	pipesfcache.SetupTokenCache(expectedConnections)
}
