package pipesfutil

import "github.com/Fajurion/pipes"

var CurrentConfig = Config{
	ExpectedConnections: 1000,
}

type Config struct {
	ExpectedConnections   int64
	NodeDisconnectHandler func(node pipes.Node)
}

func RemoveString(slice []string, s string) []string {
	for i, v := range slice {
		if v == s {
			return append(slice[:i], slice[i+1:]...)
		}
	}

	return slice
}
