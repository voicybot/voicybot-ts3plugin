package youtubedl

import (
	"net/url"

	"github.com/voicybot/voicybot-ts3plugin/resolvers"

	"github.com/BrianAllred/goydl"
)

type YoutubeDLResolver struct {
}

func (resolver *YoutubeDLResolver) Id() string {
	return pluginId
}

func (resolver *YoutubeDLResolver) DisplayName() string {
	return resolver.Id()
}

func (resolver *YoutubeDLResolver) ResolveURL(uri *url.URL, videoPassword string) (result *resolvers.ResolveResult, err error) {
	ydl := goydl.NewYoutubeDl()
	ydl.VideoURL = uri.String()
	ydl.Options.AbortOnError.Value = true
	ydl.Options.VideoPassword.Value = videoPassword
	ydl.Options.Format.Value = "bestaudio/best"
	ydl.Options.Output.Value = "-"
	ydl.Options.Quiet.Value = true

	_, err = ydl.Download()
	if err != nil {
		return
	}

	streamOutput := ydl.Stdout
	errorOutput := ydl.Stderr

	result = &resolvers.ResolveResult{
		StreamOutput: streamOutput,
		ErrorOutput:  errorOutput,
		Title:        ydl.Info.Title,
	}
	return
}
