package balancer

import (
	"errors"
	"github.com/stathat/consistent"
	"math/rand"
	"sync/atomic"
)

type Balancer interface {
	Next(key string) (string, error)
}

type RoundRobinBalancer struct {
	servers []string
	index   uint32
}

func NewRoundRobinBalancer(servers []string) *RoundRobinBalancer {
	return &RoundRobinBalancer{
		servers: servers,
		index:   rand.Uint32(),
	}
}

func (robin *RoundRobinBalancer) Next(key string) (string, error) {
	if len(robin.servers) == 0 {
		return "", errors.New("no available servers")
	}
	idx := atomic.AddUint32(&robin.index, 1)
	return robin.servers[(idx)%uint32(len(robin.servers))], nil
}

type RandomBalancer struct {
	servers []string
}

func NewRandomBalancer(servers []string) *RandomBalancer {
	return &RandomBalancer{
		servers: servers,
	}
}

func (rb *RandomBalancer) Next(key string) (string, error) {
	if len(rb.servers) == 0 {
		return "", errors.New("no available servers")
	}
	return rb.servers[rand.Intn(len(rb.servers))], nil
}

type ConsistentHashBalancer struct {
	hashRing *consistent.Consistent
}

func NewConsistentHashBalancer(servers []string, virtualNodes int) *ConsistentHashBalancer {
	ch := consistent.New()
	ch.NumberOfReplicas = virtualNodes
	for _, server := range servers {
		ch.Add(server)
	}
	return &ConsistentHashBalancer{
		hashRing: ch,
	}
}

func (ch *ConsistentHashBalancer) Next(key string) (string, error) {
	if ch.hashRing == nil {
		return "", errors.New("consistent hash ring is not initialized")
	}
	return ch.hashRing.Get(key)
}
