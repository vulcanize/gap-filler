package mux

import (
	"net/url"

	"github.com/ethereum/go-ethereum/rpc"
)

// Options configurations for proxy service
type Options struct {
	PostgraphileAddr *url.URL
	BasePath         string
	EnableGraphiQL   bool
	RPCClient        *rpc.Client
}
