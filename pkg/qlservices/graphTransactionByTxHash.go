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
	log.Debug("do request to geth")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	var data json.RawMessage
	log.Debug("call debug_writeTxTraceGraph")

	// since the clients are not being modified after initialization, it is safe to iterate over this list in separate goroutines
	for _, client := range srv.clients {
		// if deadline has been reached, break
		// otherwise it'd keep calling the rest of the clients with the exhausted deadline
		select {
		case <-ctx.Done():
			return DeadlineReached
		default:
		}
		err = client.CallContext(ctx, &data, "debug_writeTxTraceGraph", hash.Hex())
		if err == nil {
			log.WithField("resp", data).Debug("debug_writeTxTraceGraph result")
			return nil
		}
		log.WithError(err).Debug("bad debug_writeTxTraceGraph request")
	}

	return err
}
