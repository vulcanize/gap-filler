package qlparser

import (
	"math/big"
	"strings"
	"testing"
)

func TestEthHeaderCidByBlockNumberArgEmptyBody(t *testing.T) {
	n, i, err := EthHeaderCidByBlockNumberArg(strings.NewReader(``))
	if n != nil {
		t.Errorf("Want: nil, Got: %v", n)
	}
	if i != nil {
		t.Errorf("Want: nil, Got: %v", i)
	}
	if err != nil {
		t.Errorf("Want: nil, Got: %v", err)
	}
}

func TestEthHeaderCidByBlockNumberArgNoQuery(t *testing.T) {
	n, i, err := EthHeaderCidByBlockNumberArg(strings.NewReader(`
		query MyQuery {
			ethHeaderCid(nodeId: "")
		}	
	`))
	if n != nil {
		t.Errorf("Want: nil, Got: %v", n)
	}
	if i != nil {
		t.Errorf("Want: nil, Got: %v", i)
	}
	if err != nil {
		t.Errorf("Want: nil, Got: %v", err)
	}
}

func TestEthHeaderCidByBlockNumberArgNoArg(t *testing.T) {
	n, i, err := EthHeaderCidByBlockNumberArg(strings.NewReader(`
		query MyQuery {
			ethHeaderCidByBlockNumber
		}	
	`))
	if n != nil {
		t.Errorf("Want: nil, Got: %v", n)
	}
	if i != nil {
		t.Errorf("Want: nil, Got: %v", i)
	}
	if err != nil {
		t.Errorf("Want: nil, Got: %v", err)
	}
}

func TestEthHeaderCidByBlockNumberArgSimple(t *testing.T) {
	queries := []string{
		`
			query MyQuery {
				ethHeaderCidByBlockNumber(n: "100000")
			}
		`,
		`
			query MyQuery {
				ethHeaderCidByBlockNumber(n: "100000") {
					edges {
						cursor
						node {
							blockHash
							blockNumber
						}
					}
				}
			}
		`,
	}
	for _, query := range queries {
		n, i, err := EthHeaderCidByBlockNumberArg(strings.NewReader(query))
		if n == nil || n.Cmp(big.NewInt(100000)) != 0 {
			t.Errorf("Want: 100000, Got: %s", n)
		}
		if i == nil || (*i != 0) {
			t.Errorf("Want: nil, Got: %v", *i)
		}
		if err != nil {
			t.Errorf("Want: nil, Got: %v", err)
		}
	}
}

func TestEthHeaderCidByBlockNumberArgMixedQueries(t *testing.T) {
	type query struct {
		Source string
		N      *big.Int
		I      int
	}
	queries := make([]query, 2)
	queries[0] = query{
		Source: `
			query MyQuery {
				blockByKey(key: "")
				ethHeaderCidByBlockNumber(n: "999") {
					edges {
						cursor
						node {
							blockHash
							blockNumber
						}
					}
				}
			}
		`,
		N: big.NewInt(999),
		I: 1,
	}
	queries[1] = query{
		Source: `
			query MyQuery {
				blockByKey(key: "")
				ethHeaderCid(nodeId: "")
				ethHeaderCidByBlockNumber(n: "555") {
					edges {
						cursor
						node {
							blockHash
							blockNumber
						}
					}
				}
			}
		`,
		N: big.NewInt(555),
		I: 2,
	}

	for j, query := range queries {
		n, i, err := EthHeaderCidByBlockNumberArg(strings.NewReader(query.Source))
		if n == nil || n.Cmp(query.N) != 0 {
			t.Errorf("[%d] Want: %s, Got: %s", j, query.N, n)
		}
		if i == nil || (*i != query.I) {
			t.Errorf("[%d] Want: %v, Got: %v", j, query.I, *i)
		}
		if err != nil {
			t.Errorf("[%d] Want: nil, Got: %v", j, err)
		}
	}
}
