package mux

import (
	"net/url"

	"github.com/ethereum/go-ethereum/rpc"
)

type PostgraphileOptions struct {
	Default    *url.URL
	TracingAPI *url.URL
}

type RPCOptions struct {
	Default *rpc.Client
	Tracing *rpc.Client
}

// Options configurations for proxy service
type Options struct {
	BasePath       string
	EnableGraphiQL bool
	Postgraphile   PostgraphileOptions
	RPC            RPCOptions
}
