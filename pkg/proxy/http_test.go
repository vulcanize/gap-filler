package proxy

import (
	"io/ioutil"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/rpc"
)

func TestEthHeaderCidByBlockNumberEmptyBody(t *testing.T) {
	proxy := NewHTTPReverseProxy(nil, nil)
	proxy.req2postgraphile = func(client *http.Client, uri *url.URL, body []byte) ([]byte, error) {
		return []byte("some response"), nil
	}

	rr := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/", strings.NewReader(`
		{"query":"query MyQuery {ethHeaderCid(nodeId: "")}","variables":null,"operationName":"MyQuery"}
	`))
	proxy.ServeHTTP(rr, r)

	body, err := ioutil.ReadAll(rr.Body)
	if err != nil {
		t.Error(err)
	}

	if string(body) != "some response" {
		t.Errorf("Want: 'some response', Got: '%s'", string(body))
	}
}

func TestEthHeaderCidByBlockNumberSimple(t *testing.T) {
	json := `
		{
			"data": {
				"ethHeaderCidByBlockNumber": {
					"edges": [
						{
							"cursor": "WyJuYXR1cmFsIiwxXQ==",
							"node": {
								"blockHash": "0xb24ca88bcc460976afd78e6887f4b94078a234d59219b523f449c2414b544c70",
								"blockNumber": "123"
							}
						}
					]
				}
			}
		}
	`

	proxy := NewHTTPReverseProxy(nil, nil)
	proxy.req2postgraphile = func(client *http.Client, uri *url.URL, body []byte) ([]byte, error) {
		return []byte(json), nil
	}

	rr := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/", strings.NewReader(`
		{"query":"query MyQuery {\n  ethHeaderCidByBlockNumber(n: \"123\") {\n    edges {\n      cursor\n      node {\n        blockHash\n        blockNumber\n      }\n    }\n  }\n}\n","variables":null,"operationName":"MyQuery"}
	`))
	proxy.ServeHTTP(rr, r)

	body, err := ioutil.ReadAll(rr.Body)
	if err != nil {
		t.Error(err)
	}

	if strings.Compare(json, string(body)) != 0 {
		t.Errorf("Want: %s, Got: '%s'", json, string(body))
	}
}

func TestEthHeaderCidByBlockNumberDataPulling(t *testing.T) {
	json := `
		{
			"data": {
				"ethHeaderCidByBlockNumber": {
					"edges": [
						{
							"cursor": "WyJuYXR1cmFsIiwxXQ==",
							"node": {
								"blockHash": "0xb24ca88bcc460976afd78e6887f4b94078a234d59219b523f449c2414b544c70",
								"blockNumber": "123"
							}
						}
					]
				}
			}
		}
	`

	proxy := NewHTTPReverseProxy(nil, nil)
	proxy.req2postgraphile = func(client *http.Client, uri *url.URL, body []byte) ([]byte, error) {
		return []byte(`
			{
				"data": {
					"ethHeaderCidByBlockNumber": {
						"edges": []
					}
				}
			}
		`), nil
	}
	isReq2statediffCalled := false
	proxy.req2statediff = func(rpc *rpc.Client, n *big.Int) error {
		isReq2statediffCalled = true
		return nil
	}
	proxy.datapuller = func(r *http.Request, body []byte) ([]byte, error) {
		return []byte(json), nil
	}

	rr := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/", strings.NewReader(`
		{"query":"query MyQuery {\n  ethHeaderCidByBlockNumber(n: \"123\") {\n    edges {\n      cursor\n      node {\n        blockHash\n        blockNumber\n      }\n    }\n  }\n}\n","variables":null,"operationName":"MyQuery"}
	`))
	proxy.ServeHTTP(rr, r)

	body, err := ioutil.ReadAll(rr.Body)
	if err != nil {
		t.Error(err)
	}

	if !isReq2statediffCalled {
		t.Error("isReq2statediffCalled were not called")
	}

	if strings.Compare(json, string(body)) != 0 {
		t.Errorf("Want: %s, Got: '%s'", json, string(body))
	}
}
