package qlservices

import (
	"context"
	"encoding/json"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/statediff"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fastjson"
)

var stateDiffMethod = "statediff_writeStateDiffAt"

type EthHeaderCidByBlockNumberService struct {
	clients []*rpc.Client
}

func NewEthHeaderCidByBlockNumberService(clients []*rpc.Client) *EthHeaderCidByBlockNumberService {
	return &EthHeaderCidByBlockNumberService{clients: clients}
}

func (srv *EthHeaderCidByBlockNumberService) Name() string {
	return "ethHeaderCidByBlockNumber"
}

func (srv *EthHeaderCidByBlockNumberService) args(args []*ast.Argument) (*big.Int, error) {
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
	_, err := srv.args(args)
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
	n, err := srv.args(args)
	if err != nil {
		return err
	}
	params := statediff.Params{
		IncludeBlock:    true,
		IncludeReceipts: true,
		IncludeTD:       true,
		IncludeCode:     true,
	}
	log := logrus.WithFields(logrus.Fields{
		"blockNum": n,
		"params":   params,
	})
	log.Debug("do request to Geth")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	var data json.RawMessage

	return proxyCallContext(srv.clients, log, ctx, &data, stateDiffMethod, n.Uint64(), params)
}
