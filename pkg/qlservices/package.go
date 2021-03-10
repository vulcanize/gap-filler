package qlservices

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/sirupsen/logrus"
)

var (
	DeadlineReached = errors.New("context deadline reached")
	ErrNoArgs       = errors.New("no arguments")
	ErrBadType      = errors.New("bad argument type")
)

func proxyCallContext(clients []*rpc.Client, log *logrus.Entry, ctx context.Context, res *json.RawMessage, method string, args ...interface{}) error {
	var err error
	log.Debugf("proxy call %s", method)
	for _, client := range clients {
		// if deadline has been reached, break
		// otherwise it'd keep calling the rest of the clients with the exhausted deadline
		select {
		case <-ctx.Done():
			return DeadlineReached
		default:
		}
		err = client.CallContext(ctx, res, method, args...)
		if err == nil {
			log.WithField("resp", *res).Debugf("%s result", method)
			return nil
		}
		log.WithError(err).Debugf("bad %s request", method)
	}
	return err
}
