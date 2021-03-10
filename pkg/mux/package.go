package mux

import (
	"net/http"
	"path"

	"github.com/friendsofgo/graphiql"
	"github.com/vulcanize/gap-filler/pkg/proxy"
)

// NewServeMux create new http service
func NewServeMux(opts *Options) (*http.ServeMux, error) {
	mux := http.NewServeMux()

	if opts.EnableGraphiQL {
		grphiql, err := graphiql.NewGraphiqlHandler(path.Join(opts.BasePath, "/graphql"))
		if err != nil {
			return nil, err
		}
		mux.Handle(path.Join(opts.BasePath, "/graphiql"), grphiql)
	}

	mux.Handle(path.Join(opts.BasePath, "/graphql"), proxy.New(&proxy.Options{
		RPC: proxy.RPCOptions{
			DefaultClients: opts.RPC.DefaultClients,
			TracingClients: opts.RPC.TracingClients,
		},
		Postgraphile: proxy.PostgraphileOptions{
			Default:    opts.Postgraphile.Default,
			TracingAPI: opts.Postgraphile.TracingAPI,
		},
	}))

	return mux, nil
}
