package proxy

import (
	"net/http"
	"net/url"

	"github.com/ethereum/go-ethereum/rpc"

	"github.com/vulcanize/gap-filler/pkg/qlservices"
)

// Proxy accept http and ws requests
type Proxy struct {
	wsProxy   http.Handler
	httpProxy http.Handler
}

type PostgraphileOptions struct {
	Default    *url.URL
	TracingAPI *url.URL
}

type RPCOptions struct {
	DefaultClients []*rpc.Client
	TracingClients []*rpc.Client
}

type Options struct {
	Postgraphile PostgraphileOptions
	RPC          RPCOptions
}

// New create new router
func New(opts *Options) *Proxy {
	return &Proxy{
		wsProxy: NewWebsocketReverseProxy(opts.Postgraphile.Default),
		httpProxy: NewHTTPReverseProxy(opts).
			Register(qlservices.NewEthHeaderCidByBlockNumberService(opts.RPC.DefaultClients)).
			Register(qlservices.NewGetGraphCallByTxHashService(opts.RPC.TracingClients)),
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var proxy http.Handler
	if IsWebSocketRequest(r) {
		proxy = p.wsProxy
	} else {
		proxy = p.httpProxy
	}
	proxy.ServeHTTP(w, r)
}
