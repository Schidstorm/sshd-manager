package putter

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/mpvl/unique"
	"github.com/schidstorm/sshd-manager/manager"
	"github.com/sirupsen/logrus"
	"sort"
	"strings"
	"time"
)

type Etcd struct {
	endpoints []string
	client    *clientv3.Client
	kv        clientv3.KV
}

func NewEtcd(config EtcdConfig) *Etcd {
	return &Etcd{
		endpoints: config.Endpoints,
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

func (receiver *Etcd) getGroupKeyListOfServer(ctx context.Context, serverKey string) []string {
	groupsPrefix := fmt.Sprintf("/servers/%s/groups/", serverKey)
	etcdGroupKVPs := receiver.getAllByPrefix(ctx, groupsPrefix)

	var groupKeys []string
	for groupKey, flag := range etcdGroupKVPs {
		if flag == "1" {
			groupKeys = append(groupKeys, groupKey)
		}
	}

	return groupKeys
}

func (receiver *Etcd) GetPublicKeysByServerKey(ctx context.Context, serverKey string) []string {
	var publicKeys []string
	for _, groupKey := range receiver.getGroupKeyListOfServer(ctx, serverKey) {
		etcdGroupAccounts := receiver.getAllByPrefix(ctx, fmt.Sprintf("/groups/%s/accounts/", groupKey))
		for _, publicKey := range etcdGroupAccounts {
			publicKeys = append(publicKeys, publicKey)
		}
	}

	sort.Strings(publicKeys)
	unique.Strings(&publicKeys)

	return publicKeys
}

func (receiver *Etcd) AddServer(ctx context.Context, server *manager.Server) {
	receiver.put(ctx, fmt.Sprintf("/servers/%s/hostname", server.Key), server.Hostname)

	etcdGroups := receiver.getAllByPrefix(ctx, fmt.Sprintf("/servers/%s/groups/", server.Key))
	for _, group := range server.Groups {
		delete(etcdGroups, group)
	}
	for deleteGroupKey, _ := range etcdGroups {
		receiver.delete(ctx, fmt.Sprintf("/servers/%s/groups/%s", server.Key, deleteGroupKey))
	}

	for _, group := range server.Groups {
		receiver.put(ctx, fmt.Sprintf("/servers/%s/groups/%s", server.Key, group), "1")
	}
}

func (receiver *Etcd) AddAccount(ctx context.Context, account *manager.Account) {
	receiver.put(ctx, fmt.Sprintf("/accounts/%s/label", account.Key), account.Label)
	receiver.put(ctx, fmt.Sprintf("/accounts/%s/publicKey", account.Key), account.PublicKey)

	etcdGroups := receiver.getAllByPrefix(ctx, "/groups/")
	for etcdGroupKey, _ := range etcdGroups {
		parts := strings.Split(etcdGroupKey, "/")
		groupKey := parts[0]
		accountKey := parts[2]
		if accountKey == account.Key && !in(account.Groups, groupKey) {
			receiver.delete(ctx, fmt.Sprintf("/groups/%s", etcdGroupKey))
		}
	}

	for _, group := range account.Groups {
		receiver.put(ctx, fmt.Sprintf("/groups/%s/accounts/%s", group, account.Key), account.PublicKey)
	}
}

func in(list []string, element string) bool {
	for _, item := range list {
		if item == element {
			return true
		}
	}

	return false
}

func (receiver *Etcd) put(ctx context.Context, key string, value string) {
	receiver.kv.Put(ctx, key, value)
}

func (receiver *Etcd) delete(ctx context.Context, key string) {
	receiver.kv.Delete(ctx, key, clientv3.WithPrefix())
}

func (receiver *Etcd) getAllByPrefix(ctx context.Context, prefix string) map[string]string {
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

func (receiver *Etcd) Disconnect() {
	if receiver.client == nil {
		return
	}
	err := receiver.client.Close()
	if err != nil {
		logrus.Warning(err)
	}
}
