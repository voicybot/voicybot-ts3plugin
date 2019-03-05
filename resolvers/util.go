package resolvers

import (
	"errors"
	"io"
	"net/url"
)

var (
	ErrIncompatibleURL = errors.New("incompatible URL")
)

func Resolve(uri *url.URL, videoPassword string) (resolvedReadCloser io.ReadCloser, resolvedErrorReadCloser io.ReadCloser, err error) {
	resolvers := GetAll()

	if len(resolvers) == 0 {
		err = errors.New("No resolvers are registered.")
		return
	}

	for _, resolver := range resolvers {
		switch r := resolver.(type) {
		case URLToReadCloserResolver:
			rOutput, rErrorOutput, rErr := r.ResolveURLToReadCloser(uri, videoPassword)
			if rErr == ErrIncompatibleURL {
				continue
			}
			if rErr != nil {
				err = rErr
				return
			}
			resolvedReadCloser = rOutput
			resolvedErrorReadCloser = rErrorOutput
			return
		}
	}

	// At this point no resolver has been found that would work with this URL.
	resolvedReadCloser = nil
	resolvedErrorReadCloser = nil
	err = nil
	return
}
