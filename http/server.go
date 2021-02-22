package http

import (
	"fmt"
	"github.com/mpvl/unique"
	"github.com/schidstorm/sshd-manager/config"
	"github.com/schidstorm/sshd-manager/manager"
	"github.com/schidstorm/sshd-manager/parser"
	"github.com/schidstorm/sshd-manager/putter"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
)

func Run(endpoint string) error {
	http.HandleFunc("/account", handleAccount)
	http.HandleFunc("/server", handleServer)
	http.HandleFunc("/serverKeys", handleServerKeys)
	return http.ListenAndServe(endpoint, nil)
}

func handleServerKeys(writer http.ResponseWriter, request *http.Request) {
	etcd := putter.NewEtcd(config.GetConfig().EtcdEndpoints)
	err := etcd.Connect()
	defer etcd.Disconnect()
	if err != nil {
		panic(err)
	}
	if request.Method == "GET" {
		serverKey := request.URL.Query().Get("key")
		groupsPrefix := fmt.Sprintf("/servers/%s/groups/", serverKey)
		etcdGroupKVPs := etcd.GetAllByPrefix(request.Context(), groupsPrefix)

		var publicKeys []string
		for groupKey, flag := range etcdGroupKVPs {
			if flag == "1" {
				etcdGroupAccounts := etcd.GetAllByPrefix(request.Context(), fmt.Sprintf("/groups/%s/accounts/", groupKey))
				for _, publicKey := range etcdGroupAccounts {
					publicKeys = append(publicKeys, publicKey)
				}
			}
		}

		sort.Strings(publicKeys)
		unique.Strings(&publicKeys)

		buffer, _ := parser.SerializeJson(publicKeys)
		writer.WriteHeader(200)
		writer.Write(buffer)
		return
	}

	writer.WriteHeader(500)
	writer.Write([]byte("Failed"))
}

func handleServer(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	etcd := putter.NewEtcd(config.GetConfig().EtcdEndpoints)
	err := etcd.Connect()
	defer etcd.Disconnect()
	if err != nil {
		panic(err)
	}
	if request.Method == "POST" {
		server := &manager.Server{}
		buffer, _ := ioutil.ReadAll(request.Body)
		err := parser.ParseYaml(server, string(buffer))
		if err != nil {
			panic(err)
		}

		etcd.Put(request.Context(), fmt.Sprintf("/servers/%s/hostname", server.Key), server.Hostname)
		for _, group := range server.Groups {
			etcd.Put(request.Context(), fmt.Sprintf("/servers/%s/groups/%s", server.Key, group), "1")
		}

		writer.WriteHeader(200)
		writer.Write([]byte("OK"))
		return
	}

	writer.WriteHeader(500)
	writer.Write([]byte("Failed"))
}

func handleAccount(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	etcd := putter.NewEtcd(config.GetConfig().EtcdEndpoints)
	err := etcd.Connect()
	defer etcd.Disconnect()
	if err != nil {
		panic(err)
	}
	if request.Method == "POST" {
		account := &manager.Account{}
		buffer, _ := ioutil.ReadAll(request.Body)
		err := parser.ParseYaml(account, string(buffer))
		if err != nil {
			panic(err)
		}

		etcd.Put(request.Context(), fmt.Sprintf("/accounts/%s/label", account.Key), account.Label)
		etcd.Put(request.Context(), fmt.Sprintf("/accounts/%s/publicKey", account.Key), account.PublicKey)

		etcdGroups := etcd.GetAllByPrefix(request.Context(), "/groups/")
		for etcdGroupKey, _ := range etcdGroups {
			parts := strings.Split(etcdGroupKey, "/")
			accountKey := parts[2]
			if accountKey == account.Key {
				etcd.Delete(request.Context(), fmt.Sprintf("/groups/%s", etcdGroupKey))
			}
		}

		for _, group := range account.Groups {
			etcd.Put(request.Context(), fmt.Sprintf("/groups/%s/accounts/%s", group, account.Key), account.PublicKey)
		}

		writer.WriteHeader(200)
		writer.Write([]byte("OK"))
		return
	}

	writer.WriteHeader(500)
	writer.Write([]byte("Failed"))
}
