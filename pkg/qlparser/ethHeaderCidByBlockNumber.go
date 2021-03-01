package qlparser

import (
	"math/big"

	"github.com/valyala/fastjson"
)

// EthHeaderCidByBlockNumberArg detect graphql query `ethHeaderCidByBlockNumber`
func EthHeaderCidByBlockNumberArg(query []byte) (*big.Int, error) {
	prm, err := GetParam(query, "ethHeaderCidByBlockNumber", "n")
	if err != nil {
		return nil, err
	}
	value, ok := prm.Value.GetValue().(string)
	if !ok {
		return nil, ErrBadType
	}
	n, ok := new(big.Int).SetString(value, 10)
	if !ok {
		return nil, ErrBadType
	}
	return n, nil
}

// IsHaveEthHeaderCidByBlockNumberData check response is not empty
func IsHaveEthHeaderCidByBlockNumberData(data []byte) (bool, error) {
	json, err := fastjson.ParseBytes(data)
	if err != nil {
		return true, err
	}

	header := json.Get("data", "ethHeaderCidByBlockNumber")
	if header == nil {
		return true, nil
	}

	// can contain nodes or header
	var arrValue *fastjson.Value
	nodes := header.Get("nodes")
	if nodes == nil {
		// check edges
		edges := header.Get("edges")
		if edges == nil {
			return true, nil
		}

		arrValue = edges
	} else {
		arrValue = nodes
	}

	aEdges, err := arrValue.Array()
	if err != nil {
		return true, err
	}

	return len(aEdges) == 0, nil
}
