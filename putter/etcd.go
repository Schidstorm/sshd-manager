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

func (receiver *Etcd) ListServers(ctx context.Context) []*manager.Server {
	etcdServersKeys := receiver.getAllByPrefix(ctx, "/servers/")
	serverKeysMap := map[string]map[string]string{}
	for etcdServerKey, v := range etcdServersKeys {
		serverKey := strings.Split(etcdServerKey, "/")[0]
		if _, ok := serverKeysMap[serverKey]; !ok {
			serverKeysMap[serverKey] = map[string]string{}
		}
		serverKeysMap[serverKey][strings.TrimPrefix(etcdServerKey, serverKey+"/")] = v
	}

	var servers []*manager.Server
	for serverKey, serverMap := range serverKeysMap {
		servers = append(servers, keysToServer(serverMap, serverKey))
	}

	return servers
}

func (receiver *Etcd) GetServer(ctx context.Context, serverKey string) *manager.Server {
	return keysToServer(receiver.getAllByPrefix(ctx, fmt.Sprintf("/servers/%s/", serverKey)), serverKey)
}

func keysToServer(data map[string]string, serverKey string) *manager.Server {
	var groups []string
	for key, val := range data {
		if strings.HasPrefix(key, "groups/") && val == "1" {
			groups = append(groups, strings.TrimPrefix(key, "groups/"))
		}
	}

	server := &manager.Server{
		Key:      serverKey,
		Hostname: mapGetOrDefaultString(data, "hostname", ""),
		Groups:   groups,
	}
	return server
}

func (receiver *Etcd) ListAccounts(ctx context.Context) []*manager.Account {
	etcdAccountsKeys := receiver.getAllByPrefix(ctx, "/accounts/")
	accountKeysMap := map[string]map[string]string{}
	for etcdAccountKey, v := range etcdAccountsKeys {
		accountKey := strings.Split(etcdAccountKey, "/")[0]
		if _, ok := accountKeysMap[accountKey]; !ok {
			accountKeysMap[accountKey] = map[string]string{}
		}
		accountKeysMap[accountKey][strings.TrimPrefix(etcdAccountKey, accountKey+"/")] = v
	}

	var accounts []*manager.Account
	for accountKey, accountMap := range accountKeysMap {
		accounts = append(accounts, keysToAccount(accountMap, accountKey))
	}

	return accounts
}

func (receiver *Etcd) GetAccount(ctx context.Context, accountKey string) *manager.Account {
	return keysToAccount(receiver.getAllByPrefix(ctx, fmt.Sprintf("/accounts/%s/", accountKey)), accountKey)
}

func keysToAccount(data map[string]string, accountKey string) *manager.Account {
	var groups []string
	for key, val := range data {
		if strings.HasPrefix(key, "groups/") && val == "1" {
			groups = append(groups, strings.TrimPrefix(key, "groups/"))
		}
	}

	account := &manager.Account{
		Key:       accountKey,
		Label:     mapGetOrDefaultString(data, "label", ""),
		PublicKey: mapGetOrDefaultString(data, "publicKey", ""),
		Groups:    groups,
	}
	return account
}

func (receiver *Etcd) DeleteAccount(ctx context.Context, accountKey string) {
	receiver.deletePrefix(ctx, fmt.Sprintf("/accounts/%s/", accountKey))
	for _, group := range receiver.listFirstLevelChildren(ctx, "/groups/") {
		receiver.deletePrefix(ctx, fmt.Sprintf("/groups/%s/accounts/%s", group, accountKey))
	}
}

func (receiver *Etcd) DeleteServer(ctx context.Context, serverKey string) {
	receiver.deletePrefix(ctx, fmt.Sprintf("/servers/%s/", serverKey))
	receiver.delete(ctx, fmt.Sprintf("/servers/%s", serverKey))
}

func (receiver *Etcd) listFirstLevelChildren(ctx context.Context, prefix string) []string {
	var children []string
	for key, _ := range receiver.getAllByPrefix(ctx, prefix) {
		children = append(children, strings.Split(key, "/")[0])
	}

	sort.Strings(children)
	unique.Strings(&children)
	return children
}

func mapGetOrDefaultString(data map[string]string, key, defaultValue string) string {
	if val, ok := data[key]; ok {
		return val
	} else {
		return defaultValue
	}
}

func (receiver *Etcd) AddServer(ctx context.Context, server *manager.Server) {
	receiver.DeleteServer(ctx, server.Key)
	receiver.put(ctx, fmt.Sprintf("/servers/%s/hostname", server.Key), server.Hostname)

	for _, group := range server.Groups {
		receiver.put(ctx, fmt.Sprintf("/servers/%s/groups/%s", server.Key, group), "1")
	}
}

func (receiver *Etcd) AddAccount(ctx context.Context, account *manager.Account) {
	receiver.DeleteAccount(ctx, account.Key)
	receiver.put(ctx, fmt.Sprintf("/accounts/%s/label", account.Key), account.Label)
	receiver.put(ctx, fmt.Sprintf("/accounts/%s", account.Key), "1")
	receiver.put(ctx, fmt.Sprintf("/accounts/%s/publicKey", account.Key), account.PublicKey)

	for _, group := range account.Groups {
		receiver.put(ctx, fmt.Sprintf("/groups/%s/accounts/%s", group, account.Key), account.PublicKey)
		receiver.put(ctx, fmt.Sprintf("/accounts/%s/groups/%s", account.Key, group), "1")
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

func (receiver *Etcd) deletePrefix(ctx context.Context, prefix string) {
	receiver.kv.Delete(ctx, prefix, clientv3.WithPrefix())
}

func (receiver *Etcd) delete(ctx context.Context, key string) {
	receiver.kv.Delete(ctx, key)
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

func (receiver *Etcd) getAll(ctx context.Context, key string) map[string]string {
	res, err := receiver.kv.Get(ctx, key)
	if err != nil {
		return map[string]string{}
	}

	kvs := map[string]string{}
	for _, kvp := range res.Kvs {
		kvs[string(kvp.Key)] = string(kvp.Value)
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
