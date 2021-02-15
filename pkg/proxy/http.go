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
	tgDefault *url.URL
	tgTracing *url.URL
	client    *http.Client
	forward   func(body []byte) ([]byte, error)
	polling   func(r *http.Request, body []byte, names []string) ([]byte, error)

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
		tgDefault:    opts.Postgraphile.Default,
		tgTracing:    opts.Postgraphile.TracingAPI,
		client:       client,
		serviceNames: make([]string, 0),
		services:     make(map[string]Service),
	}
	proxy.forward = func(body []byte) ([]byte, error) {
		req, err := http.NewRequest("POST", proxy.tgDefault.String(), bytes.NewReader(body))
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
	proxy.polling = func(r *http.Request, body []byte, names []string) ([]byte, error) {
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

				data, err := proxy.forward(body)
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

func (handler *HTTPReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	logrus.WithField("body", string(reqBody)).Debug("new request")

	data, err := handler.forward(reqBody)
	if err != nil {
		logrus.WithError(err).Error("postgraphile first request request")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	logrus.WithField("data", string(data)).Debug("postgraphile first request request")

	args, err := qlparser.QueryParams(fastjson.GetBytes(reqBody, "query"), handler.serviceNames)
	if err != nil {
		logrus.WithError(err).Error("can't parse graphQL queries and params")
		w.Write(data)
		return
	}

	emptyQueries := make([]string, 0)
	for query := range args {
		log := logrus.WithField("service", query)
		if err := handler.services[query].Validate(args[query]); err != nil {
			log.WithError(err).Error("bad arguments")
			w.Write(data)
			return
		}
		empty, err := handler.services[query].IsEmpty(data)
		if err != nil {
			log.WithError(err).Errorf("can't call %s.IsEmpty", query)
			w.Write(data)
			return
		}
		if empty {
			emptyQueries = append(emptyQueries, query)
		}
	}

	if len(emptyQueries) == 0 {
		w.Write(data)
		return
	}

	for _, query := range emptyQueries {
		if err := handler.services[query].Do(args[query]); err != nil {
			logrus.WithError(err).Errorf("can't call %s.Do", query)
			w.Write(data)
			return
		}
	}

	polledData, err := handler.polling(r, reqBody, emptyQueries)
	if err != nil {
		logrus.WithError(err).Error("have error after polling")
		w.Write(data)
		return
	}

	w.Write(polledData)
}
