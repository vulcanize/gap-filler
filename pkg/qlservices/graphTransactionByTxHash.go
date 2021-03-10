package qlservices

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fastjson"
)

var traceMethod = "debug_writeTxTraceGraph"

type GraphTransactionByTxHashService struct {
	number  int
	clients []*rpc.Client
}

func NewGetGraphCallByTxHashService(clients []*rpc.Client) *GraphTransactionByTxHashService {
	return &GraphTransactionByTxHashService{clients: clients}
}

func (srv *GraphTransactionByTxHashService) Name() string {
	return "graphTransactionByTxHash"
}

func (srv *GraphTransactionByTxHashService) params(args []*ast.Argument) (common.Hash, error) {
	if len(args) == 0 {
		return common.Hash{}, ErrNoArgs
	}
	value, ok := args[0].Value.GetValue().(string)
	if !ok {
		return common.Hash{}, ErrBadType
	}
	return common.HexToHash(value), nil
}

func (srv *GraphTransactionByTxHashService) Validate(args []*ast.Argument) error {
	_, err := srv.params(args)
	return err
}

func (srv *GraphTransactionByTxHashService) IsEmpty(data []byte) (bool, error) {
	json, err := fastjson.ParseBytes(data)
	if err != nil {
		return true, err
	}

	header := json.Get("data", "graphTransactionByTxHash")
	return header == nil || header.Type() == fastjson.TypeNull, nil
}

func (srv *GraphTransactionByTxHashService) Do(args []*ast.Argument) error {
	hash, err := srv.params(args)
	if err != nil {
		return err
	}
	log := logrus.WithField("hash", hash.Hex())
	log.Debug("do request to Geth")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	var data json.RawMessage

	return proxyCallContext(srv.clients, log, ctx, &data, traceMethod, hash.Hex())
}
