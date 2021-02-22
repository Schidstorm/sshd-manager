package putter

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"strings"
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
		Endpoints:   receiver.endpoints,
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

func (receiver *Etcd) Delete(ctx context.Context, key string) {
	receiver.kv.Delete(ctx, key)
}

func (receiver *Etcd) GetAllByPrefix(ctx context.Context, prefix string) map[string]string {
	res, err := receiver.kv.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return map[string]string{}
	}

	kvs := map[string]string{}
	for _, kvp := range res.Kvs {
		key := strings.TrimPrefix(string(kvp.Key), prefix)
		kvs[key] = string(kvp.Value)
	}
	return kvs
}

func (receiver *Etcd) Disconnect() error {
	return receiver.client.Close()
}
