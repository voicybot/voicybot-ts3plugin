package resolvers

import (
	"errors"
	"net/url"
)

var (
	ErrIncompatibleURL = errors.New("incompatible URL")
)

func Resolve(uri *url.URL, videoPassword string) (result *ResolveResult, err error) {
	resolvers := GetAll()

	if len(resolvers) == 0 {
		err = errors.New("No resolvers are registered.")
		return
	}

	for _, resolver := range resolvers {
		rResult, rErr := resolver.ResolveURL(uri, videoPassword)
		if rErr == ErrIncompatibleURL {
			continue
		}
		if rErr != nil {
			err = rErr
			return
		}
		result = rResult
		return
	}

	// At this point no resolver has been found that would work with this URL.
	result = nil
	err = nil
	return
}
