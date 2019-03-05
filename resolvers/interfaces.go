package resolvers

import (
	"net/url"
)

type Resolver interface {
	Id() string
	DisplayName() string
	ResolveURL(uri *url.URL, videoPassword string) (result *ResolveResult, err error)
}
