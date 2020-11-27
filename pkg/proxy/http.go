package proxy

import (
	"io"
	"net/http"
)

// HTTPProxy proxy handler
type HTTPProxy struct {
	http.Handler

	addr   string
	client *http.Client
}

// NewHTTPProxy create new http-proxy-handler
func NewHTTPProxy(addr string) *HTTPProxy {
	return &HTTPProxy{
		addr:   addr,
		client: new(http.Client),
	}
}

func (handler *HTTPProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest("POST", handler.addr, r.Body)
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
