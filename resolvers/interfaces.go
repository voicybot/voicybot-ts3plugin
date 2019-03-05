package resolvers

import (
	"io"
	"net/url"
)

type Resolver interface {
	Id() string
	DisplayName() string
}

type URLToReadCloserResolver interface {
	Resolver
	ResolveURLToReadCloser(uri *url.URL, videoPassword string) (streamOutput io.ReadCloser, errorOutput io.ReadCloser, err error)
}
