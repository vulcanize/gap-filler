package qlservices

import (
	"errors"
	"github.com/ethereum/go-ethereum/rpc"
	"sync"
)

var (
	//ErrNoAvailableItem no item is available
	ErrNoAvailableItem = errors.New("no item is available")
)

// Balancer is an interface for representing round-robin for JSON-RPC client
type Balancer interface {
	Next() *rpc.Client
}

// Balancer RoundRobin instance
type PRCBalancer struct {
	m sync.Mutex

	index int
	items []*rpc.Client
}

func NewBalancer(items []*rpc.Client) (Balancer, error) {
	if len(items) == 0 {
		return nil, ErrNoAvailableItem
	}

	return &PRCBalancer{
		items: items,
	}, nil
}

// Next returns index address
func (b *PRCBalancer) Next() *rpc.Client {
	b.m.Lock()
	defer b.m.Unlock()

	item := b.items[b.index]
	b.index = (b.index + 1) % len(b.items)

	return item
}
