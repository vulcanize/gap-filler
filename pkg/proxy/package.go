package proxy

import (
	"net/http"
	"net/url"
)

// Proxy accept http and ws requests
type Proxy struct {
	addr *url.URL

	wsProxy *WebsocketReverseProxy
	htProxy *HTTPReverseProxy
}

// New create new router
func New(addr *url.URL) *Proxy {
	return &Proxy{
		addr:    addr,
		wsProxy: NewWebsocketReverseProxy(addr),
		htProxy: NewHTTPReverseProxy(addr),
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var proxy http.Handler
	if IsWebSocketRequest(r) {
		proxy = p.wsProxy
	} else {
		proxy = p.htProxy
	}
	proxy.ServeHTTP(w, r)
}
