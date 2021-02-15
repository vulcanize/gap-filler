package proxy

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/vulcanize/gap-filler/pkg/qlservices"
)

type EthHeaderCidByBlockNumberMockService struct {
	*qlservices.EthHeaderCidByBlockNumberService
	DoCalled bool
}

func NewEthHeaderCidByBlockNumberMockService() *EthHeaderCidByBlockNumberMockService {
	return &EthHeaderCidByBlockNumberMockService{new(qlservices.EthHeaderCidByBlockNumberService), false}
}

func (srv *EthHeaderCidByBlockNumberMockService) Do(args []*ast.Argument) error {
	srv.DoCalled = true
	return nil
}

func TestEthHeaderCidByBlockNumberEmptyBody(t *testing.T) {
	proxy := NewHTTPReverseProxy(nil)
	proxy.forward = func(body []byte) ([]byte, error) {
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
	proxy := NewHTTPReverseProxy(nil)
	proxy.forward = func(body []byte) ([]byte, error) {
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

	proxy := NewHTTPReverseProxy(nil)
	servi := NewEthHeaderCidByBlockNumberMockService()
	proxy.Register(servi)
	proxy.forward = func(body []byte) ([]byte, error) {
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

	proxy.polling = func(r *http.Request, body []byte, names []string) ([]byte, error) {
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

	if !servi.DoCalled {
		t.Error("isReq2statediffCalled were not called")
	}

	if strings.Compare(json, string(body)) != 0 {
		t.Errorf("Want: %s, Got: '%s'", json, string(body))
	}
}
