package proxy

import (
	"net/http"
	"net/url"

	"github.com/ethereum/go-ethereum/rpc"
)

// Proxy accept http and ws requests
type Proxy struct {
	addr *url.URL

	wsProxy   http.Handler
	httpProxy http.Handler
}

// New create new router
func New(addr *url.URL, rpc *rpc.Client) *Proxy {
	return &Proxy{
		addr:      addr,
		wsProxy:   NewWebsocketReverseProxy(addr),
		httpProxy: NewHTTPReverseProxy(addr, rpc),
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
