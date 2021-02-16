package proxy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fastjson"
	"github.com/vulcanize/gap-filler/pkg/qlparser"
)

type Service interface {
	Name() string
	Validate(args []*ast.Argument) error
	IsEmpty(data []byte) (bool, error)
	Do(args []*ast.Argument) error
}

// HTTPReverseProxy it work with a regular HTTP request
type HTTPReverseProxy struct {
	pqlDefault   *url.URL
	pqlTracing   *url.URL
	client       *http.Client
	forward      func(uri *url.URL, body []byte) ([]byte, error)
	polling      func(r *http.Request, uri *url.URL, body []byte, names []string) ([]byte, error)
	mu           sync.Mutex
	serviceNames []string
	services     map[string]Service
}

// NewHTTPReverseProxy create new http-proxy-handler
func NewHTTPReverseProxy(opts *Options) *HTTPReverseProxy {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	proxy := HTTPReverseProxy{
		pqlDefault:   opts.Postgraphile.Default,
		pqlTracing:   opts.Postgraphile.TracingAPI,
		client:       client,
		serviceNames: make([]string, 0),
		services:     make(map[string]Service),
	}
	proxy.forward = func(uri *url.URL, body []byte) ([]byte, error) {
		req, err := http.NewRequest("POST", uri.String(), bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")

		res, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		data, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		return data, nil
	}
	proxy.polling = func(r *http.Request, uri *url.URL, body []byte, names []string) ([]byte, error) {
		type response struct {
			data []byte
			err  error
		}
		datach := make(chan response, 1)
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		go func(ch chan response, tick *time.Ticker) {
			for now := range tick.C {
				log := logrus.WithField("ticker", now)
				log.Debug("trying to pull data")

				data, err := proxy.forward(uri, body)
				if err != nil {
					log.WithError(err).Debug("have error after request to postgql")
					ch <- response{err: err}
					return
				}

				isEmpty := false
				for _, name := range names {
					empty, err := proxy.services[name].IsEmpty(data)
					if err != nil {
						log.WithError(err).Debug("have error response parsing")
						ch <- response{err: err}
						return
					}
					isEmpty = isEmpty || empty
				}
				if !isEmpty {
					log.WithField("data", string(data)).Debug("have some response")
					ch <- response{data: data}
					return
				}
			}
		}(datach, ticker)

		select {
		case <-r.Context().Done():
			return nil, nil
		case <-time.After(15 * time.Second):
			return nil, fmt.Errorf("polling timeout")
		case resp := <-datach:
			return resp.data, resp.err
		}
	}
	return &proxy
}

// Register new service
func (handler *HTTPReverseProxy) Register(srv Service) *HTTPReverseProxy {
	handler.mu.Lock()
	defer handler.mu.Unlock()

	handler.serviceNames = append(handler.serviceNames, srv.Name())
	handler.services[srv.Name()] = srv
	return handler
}

func (handler *HTTPReverseProxy) getPQLURI(name string) *url.URL {
	if name == "getGraphCallByTxHash" {
		return handler.pqlTracing
	}
	return handler.pqlDefault
}

func (handler *HTTPReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	ddoc, docs, err := qlparser.QuerySplit(reqBody, handler.serviceNames)

	var data []byte
	if ddoc == nil {
		o := new(fastjson.Arena).NewObject()
		o.Set("data", new(fastjson.Arena).NewObject())
		data = []byte(o.String())
	} else {
		tmp, err := handler.forward(handler.pqlDefault, ddoc)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		data = tmp
	}

	parts := make(map[string][]byte)
	for name := range docs {
		uri := handler.getPQLURI(name)
		tmp, err := handler.forward(uri, docs[name])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		parts[name] = tmp
	}

	params := make(map[string][]*ast.Argument)
	for name := range docs {
		prms, err := qlparser.GetParams(fastjson.GetBytes(docs[name], "query"), name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		params[name] = prms
	}

	wg := new(sync.WaitGroup)
	for name := range docs {
		isEmpty, _ := handler.services[name].IsEmpty(parts[name])
		if !isEmpty {
			continue
		}
		wg.Add(1)
		go func(wg *sync.WaitGroup, doc []byte, name string, args []*ast.Argument) {
			defer wg.Done()
			if err := handler.services[name].Do(args); err != nil {
				return
			}
			uri := handler.getPQLURI(name)
			tmp, err := handler.polling(r, uri, doc, []string{name})
			if err == nil {
				parts[name] = tmp
			}
		}(wg, docs[name], name, params[name])
	}
	wg.Wait()

	common := fastjson.MustParseBytes(data)
	for name := range parts {
		part := fastjson.MustParseBytes(parts[name])
		data := common.Get("data")
		data.Set(name, part.Get("data", name))
		common.Set("data", data)
	}

	w.Write([]byte(common.String()))
}
