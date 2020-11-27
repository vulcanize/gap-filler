package mux

import (
	"net/http"

	"github.com/vulcanize/gap-filler-service/pkg/proxy"
)

// NewServeMux create new http service
func NewServeMux(opts *Options) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/graphql", proxy.New(opts.PostgraphileAddr))
	return mux
}
