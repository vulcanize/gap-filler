package mux

import "net/url"

// Options configurations for proxy service
type Options struct {
	PostgraphileAddr *url.URL
}
