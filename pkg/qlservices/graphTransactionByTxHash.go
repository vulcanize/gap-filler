package qlservices

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fastjson"
)

type GraphTransactionByTxHashService struct {
	balancer Balancer
}

func NewGetGraphCallByTxHashService(balancer Balancer) *GraphTransactionByTxHashService {
	return &GraphTransactionByTxHashService{balancer}
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
	log.Debug("do request to geth")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var data json.RawMessage
	log.Debug("call debug_writeTxTraceGraph")
	if err := srv.balancer.Next().CallContext(ctx, &data, "debug_writeTxTraceGraph", hash.Hex()); err != nil {
		log.WithError(err).Debug("bad debug_writeTxTraceGraph request")
		return err
	}
	log.WithField("resp", data).Debug("debug_writeTxTraceGraph result")
	return nil
}
