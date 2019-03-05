package resolvers

import "io"

type ResolveResult struct {
	StreamOutput io.ReadCloser
	ErrorOutput  io.ReadCloser
	Title        string
}
