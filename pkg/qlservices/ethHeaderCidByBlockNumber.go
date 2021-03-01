package qlservices

import (
	"context"
	"encoding/json"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/statediff"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fastjson"
)

type EthHeaderCidByBlockNumberService struct {
	balancer Balancer
}

func NewEthHeaderCidByBlockNumberService(balancer Balancer) *EthHeaderCidByBlockNumberService {
	return &EthHeaderCidByBlockNumberService{balancer}
}

func (srv *EthHeaderCidByBlockNumberService) Name() string {
	return "ethHeaderCidByBlockNumber"
}

func (srv *EthHeaderCidByBlockNumberService) params(args []*ast.Argument) (*big.Int, error) {
	if len(args) == 0 {
		return nil, ErrNoArgs
	}
	value, ok := args[0].Value.GetValue().(string)
	if !ok {
		return nil, ErrBadType
	}
	n, ok := new(big.Int).SetString(value, 10)
	if !ok {
		return nil, ErrBadType
	}
	return n, nil
}

func (srv *EthHeaderCidByBlockNumberService) Validate(args []*ast.Argument) error {
	_, err := srv.params(args)
	return err
}

func (srv *EthHeaderCidByBlockNumberService) IsEmpty(data []byte) (bool, error) {
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

func (srv *EthHeaderCidByBlockNumberService) Do(args []*ast.Argument) error {
	n, err := srv.params(args)
	if err != nil {
		return err
	}
	logrus.WithField("blockNum", n).Debug("do request to geth")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var data json.RawMessage
	params := statediff.Params{
		IntermediateStateNodes:   true,
		IntermediateStorageNodes: true,
		IncludeBlock:             true,
		IncludeReceipts:          true,
		IncludeTD:                true,
		IncludeCode:              true,
	}
	log := logrus.WithFields(logrus.Fields{
		"n":      n,
		"params": params,
	})
	log.Debug("call statediff_stateDiffAt")
	if err := srv.balancer.Next().CallContext(ctx, &data, "statediff_writeStateDiffAt", n.Uint64(), params); err != nil {
		log.WithError(err).Debug("bad statediff_writeStateDiffAt request")
		return err
	}
	log.WithField("resp", data).Debug("statediff_writeStateDiffAt result")
	return nil
}
