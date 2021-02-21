package putter

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"time"
)

type Etcd struct {
	endpoints []string
	client    *clientv3.Client
	kv        clientv3.KV
}

func NewEtcd(endpoints []string) *Etcd {
	return &Etcd{
		endpoints: endpoints,
	}
}

func (receiver *Etcd) Connect() error {
	var err error = nil
	receiver.client, err = clientv3.New(clientv3.Config{
		DialTimeout: 10 * time.Second,
		Endpoints:   []string{"127.0.0.1:2379"},
	})
	if err != nil {
		return err
	}

	receiver.kv = clientv3.NewKV(receiver.client)
	return nil
}

func (receiver *Etcd) Put(ctx context.Context, key string, value string) {
	receiver.kv.Put(ctx, key, value)
}

func (receiver *Etcd) Disconnect() error {
	return receiver.client.Close()
}
