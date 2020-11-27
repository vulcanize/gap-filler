package proxy

import (
	"io"
	"net/http"
	"net/url"
)

// HTTPReverseProxy it work with a regular HTTP request
type HTTPReverseProxy struct {
	http.Handler

	addr   *url.URL
	client *http.Client
}

// NewHTTPReverseProxy create new http-proxy-handler
func NewHTTPReverseProxy(addr *url.URL) *HTTPReverseProxy {
	return &HTTPReverseProxy{
		addr:   addr,
		client: new(http.Client),
	}
}

func (handler *HTTPReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest("POST", handler.addr.String(), r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	res, err := handler.client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), res.StatusCode)
		return
	}
	io.Copy(w, res.Body)
}
