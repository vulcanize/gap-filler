package qlservices

import (
	"github.com/ethereum/go-ethereum/rpc"
	"sync"
	"testing"
)

func TestRoundrobin(t *testing.T) {
	addresses := []string{
		"http://127.0.0.1:8545",
		"http://127.0.0.2:8545",
		"http://127.0.0.3:8545",
		"http://127.0.0.4:8545",
		"http://127.0.0.5:8545",
	}
	clients := make([]*rpc.Client, 0, len(addresses))
	for _, address := range addresses {
		rpcClient, err := rpc.Dial(address)
		if err != nil {
			t.Error(err)
		}

		clients = append(clients, rpcClient)
	}
	mu := new(sync.Mutex)
	wg := sync.WaitGroup{}
	balancer, err := NewBalancer(clients)
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			balancer.Next()
			mu.Lock()
			defer mu.Unlock()
			wg.Done()
		}()
	}
	wg.Wait()
}
