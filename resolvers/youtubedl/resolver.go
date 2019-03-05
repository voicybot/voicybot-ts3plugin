package youtubedl

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os/exec"
	"strings"

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

func (resolver *YoutubeDLResolver) ResolveURLToReadCloser(uri *url.URL, videoPassword string) (streamOutput io.ReadCloser, errorOutput io.ReadCloser, err error) {
	ydl := goydl.NewYoutubeDl()
	ydl.VideoURL = uri.String()
	ydl.Options.AbortOnError.Value = true
	ydl.Options.VideoPassword.Value = videoPassword
	ydl.Options.Format.Value = "bestaudio/best"
	ydl.Options.Output.Value = "-"
	ydl.Options.Quiet.Value = true

	var proc *exec.Cmd
	proc, err = ydl.Download()
	if err != nil {
		return
	}

	streamOutput = ydl.Stdout
	errorOutput = ioutil.NopCloser(
		io.MultiReader(
			strings.NewReader(fmt.Sprintf("%s %+v\n", proc.Path, proc.Args)),
			ydl.Stderr,
		))
	return
}
