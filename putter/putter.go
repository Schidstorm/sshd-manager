package putter

import (
	"context"
	"github.com/schidstorm/sshd-manager/manager"
)

type PutterType string

const PutterTypeEtcd = "etcd"

type Putter interface {
	Connect() error
	Disconnect()
	AddAccount(ctx context.Context, account *manager.Account)
	AddServer(ctx context.Context, server *manager.Server)
	GetPublicKeysByServerKey(ctx context.Context, serverKey string) []string
	ListServers(ctx context.Context) []*manager.Server
	GetServer(ctx context.Context, serverKey string) *manager.Server
	ListAccounts(ctx context.Context) []*manager.Account
	GetAccount(ctx context.Context, accountKey string) *manager.Account
	DeleteAccount(ctx context.Context, accountKey string)
	DeleteServer(ctx context.Context, serverKey string)
}
