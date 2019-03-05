package youtubedl

import (
	"github.com/voicybot/voicybot-ts3plugin/resolvers"
)

const (
	pluginId = "youtube-dl"
)

func init() {
	RegisterResolver()
}

func RegisterResolver() {
	var r resolvers.Resolver = new(YoutubeDLResolver)
	resolvers.Register(r)
}

func UnregisterResolver() {
	resolvers.Unregister(resolvers.ById(pluginId))
}
