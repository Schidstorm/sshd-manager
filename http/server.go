package http

import (
	"fmt"
	"github.com/schidstorm/sshd-manager/config"
	"github.com/schidstorm/sshd-manager/manager"
	"github.com/schidstorm/sshd-manager/parser"
	"github.com/schidstorm/sshd-manager/putter"
	"io/ioutil"
	"net/http"
)

func Run(endpoint string) error {
	http.HandleFunc("/account", handleAccount)
	http.HandleFunc("/group", handleGroup)
	http.HandleFunc("/server", handleServer)
	return http.ListenAndServe(endpoint, nil)
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

func handleGroup(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	etcd := putter.NewEtcd(config.GetConfig().EtcdEndpoints)
	err := etcd.Connect()
	defer etcd.Disconnect()
	if err != nil {
		panic(err)
	}
	if request.Method == "POST" {
		group := &manager.Group{}
		buffer, _ := ioutil.ReadAll(request.Body)
		err := parser.ParseYaml(group, string(buffer))
		if err != nil {
			panic(err)
		}

		etcd.Put(request.Context(), fmt.Sprintf("/groups/%s", group.Key), "1")

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
		for _, group := range account.Groups {
			etcd.Put(request.Context(), fmt.Sprintf("/accounts/%s/groups/%s", account.Key, group), "1")
		}

		writer.WriteHeader(200)
		writer.Write([]byte("OK"))
		return
	}

	writer.WriteHeader(500)
	writer.Write([]byte("Failed"))
}
