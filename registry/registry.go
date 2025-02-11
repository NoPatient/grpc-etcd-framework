package registry

import (
	"context"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type Registry struct {
	client *clientv3.Client
	ttl    int64
}

func NewRegistry(endpoints []string, ttl int64) (*Registry, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	return &Registry{client: cli, ttl: ttl}, nil
}

func (r *Registry) Register(serviceName, addr string) error {
	leaseResp, err := r.client.Grant(context.Background(), r.ttl)
	if err != nil {
		return err
	}

	_, err = r.client.Put(context.Background(), "/services/"+serviceName+"/"+addr, addr, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}

	ch, err := r.client.KeepAlive(context.Background(), leaseResp.ID)
	if err != nil {
		return err
	}

	go func() {
		for {
			<-ch
		}
	}()

	return nil
}

func (r *Registry) Discover(serviceName string) ([]string, error) {
	resp, err := r.client.Get(context.Background(), "/services/"+serviceName, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	var addrs []string
	for _, kv := range resp.Kvs {
		addrs = append(addrs, string(kv.Value))
	}

	return addrs, nil
}
