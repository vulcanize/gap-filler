package proxy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/url"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/valyala/fastjson"
	"github.com/vulcanize/gap-filler-service/pkg/qlparser"
)

// HTTPReverseProxy it work with a regular HTTP request
type HTTPReverseProxy struct {
	target *url.URL
	client *http.Client
}

// NewHTTPReverseProxy create new http-proxy-handler
func NewHTTPReverseProxy(target *url.URL) *HTTPReverseProxy {
	return &HTTPReverseProxy{
		target: target,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (handler *HTTPReverseProxy) doReqToPostgraphile(body []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", handler.target.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := handler.client.Do(req)
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

func (handler *HTTPReverseProxy) isEmptyData(rawJSON []byte) (bool, error) {
	json, err := fastjson.ParseBytes(rawJSON)
	if err != nil {
		return false, err
	}

	edges := json.Get("data", "ethHeaderCidByBlockNumber", "edges")
	if edges == nil {
		return true, nil
	}

	aEdges, err := edges.Array()
	if err != nil {
		return false, err
	}

	return len(aEdges) == 0, nil
}

func (handler *HTTPReverseProxy) doReqToGethStateDiff(n *big.Int) error {
	logrus.WithField("blockNum", n).Debug("do request to geth")
	return nil
}

func (handler *HTTPReverseProxy) pullData(body []byte) ([]byte, error) {
	type response struct {
		data []byte
		err  error
	}
	datach := make(chan response, 1)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	go func(ch chan response, tick *time.Ticker) {
		for now := range tick.C {
			logrus.WithField("ticker", now).Debug("trying to pull data")
			data, err := handler.doReqToPostgraphile(body)
			if err != nil {
				logrus.WithField("now", now).WithError(err).Debug("have error after request to postgql")
				ch <- response{err: err}
				return
			}
			isEmpty, err := handler.isEmptyData(data)
			if err != nil {
				logrus.WithField("now", now).WithError(err).Debug("have error response parsing")
				ch <- response{err: err}
				return
			}
			if !isEmpty {
				logrus.WithField("now", now).WithField("data", string(data)).Debug("have some response")
				ch <- response{data: data}
				return
			}
		}
	}(datach, ticker)

	select {
	case <-time.After(15 * time.Second):
		return nil, fmt.Errorf("pooling timeout")
	case resp := <-datach:
		return resp.data, resp.err
	}
}

func (handler *HTTPReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	logrus.WithField("body", string(body)).Debug("new request")

	data, err := handler.doReqToPostgraphile(body)
	if err != nil {
		logrus.WithError(err).Debug("postgraphile first request request")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	logrus.WithField("data", string(data)).Debug("postgraphile first request request")

	blockNum, err := qlparser.EthHeaderCidByBlockNumberArg(fastjson.GetBytes(body, "query"))
	if err != nil {
		logrus.WithError(err).Warn("can't parse graphQL body")
	}
	if blockNum == nil {
		logrus.Debug("no block number in request")
		w.Write(data)
		return
	}

	isEmpty, err := handler.isEmptyData(data)
	if err != nil {
		logrus.WithError(err).Debug("can't check data")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !isEmpty {
		logrus.WithField("data", string(data)).Debug("data have a some body")
		w.Write(data)
		return
	}

	handler.doReqToGethStateDiff(blockNum)
	data, err = handler.pullData(body)
	if err != nil {
		logrus.WithError(err).Debug("have error after pulling")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Write(data)
}
