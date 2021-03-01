package mux

import (
	"github.com/vulcanize/gap-filler/pkg/qlservices"
	"net/url"
)

type PostgraphileOptions struct {
	Default    *url.URL
	TracingAPI *url.URL
}

type RPCOptions struct {
	DefaultBalancer qlservices.Balancer
	TracingBalancer qlservices.Balancer
}

// Options configurations for proxy service
type Options struct {
	BasePath       string
	EnableGraphiQL bool
	Postgraphile   PostgraphileOptions
	RPC            RPCOptions
}
