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

type GetGraphCallByTxHashService struct {
	rpc *rpc.Client
}

func NewGetGraphCallByTxHashService(rpc *rpc.Client) *GetGraphCallByTxHashService {
	return &GetGraphCallByTxHashService{rpc}
}

func (srv *GetGraphCallByTxHashService) Name() string {
	return "getGraphCallByTxHash"
}

func (srv *GetGraphCallByTxHashService) params(args []*ast.Argument) (common.Hash, error) {
	if len(args) == 0 {
		return common.Hash{}, ErrNoArgs
	}
	value, ok := args[0].Value.GetValue().(string)
	if !ok {
		return common.Hash{}, ErrBadType
	}
	return common.HexToHash(value), nil
}

func (srv *GetGraphCallByTxHashService) Validate(args []*ast.Argument) error {
	_, err := srv.params(args)
	return err
}

func (srv *GetGraphCallByTxHashService) IsEmpty(data []byte) (bool, error) {
	json, err := fastjson.ParseBytes(data)
	if err != nil {
		return true, err
	}

	header := json.Get("data", "getGraphCallByTxHash")
	if header == nil {
		return true, nil
	}

	return false, nil
}

func (srv *GetGraphCallByTxHashService) Do(args []*ast.Argument) error {
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
	if err := srv.rpc.CallContext(ctx, &data, "debug_writeTxTraceGraph", hash.Hex()); err != nil {
		log.WithError(err).Debug("bad debug_writeTxTraceGraph request")
		return err
	}
	log.WithField("resp", data).Debug("debug_writeTxTraceGraph result")
	return nil
}
