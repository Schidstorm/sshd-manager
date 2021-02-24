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
}
