package http

import (
	"github.com/schidstorm/sshd-manager/manager"
	"github.com/schidstorm/sshd-manager/parser"
	"github.com/schidstorm/sshd-manager/putter"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

type Server struct {
	putterObj putter.Putter
	addr      string
}

func NewServer(addr string, putterObj putter.Putter) *Server {
	return &Server{
		putterObj: putterObj,
		addr:      addr,
	}
}

func (server *Server) Run(addr string) error {
	http.HandleFunc("/account", server.handleAccount)
	http.HandleFunc("/server", server.handleServer)
	http.HandleFunc("/serverKeys", server.handleServerKeys)

	logrus.Infof("listening on %s", addr)
	return http.ListenAndServe(addr, nil)
}

func (server *Server) handleServerKeys(writer http.ResponseWriter, request *http.Request) {
	err := server.putterObj.Connect()
	defer server.putterObj.Disconnect()
	if err != nil {
		panic(err)
	}
	if request.Method == "GET" {
		serverKey := request.URL.Query().Get("key")
		publicKeys := server.putterObj.GetPublicKeysByServerKey(request.Context(), serverKey)

		buffer, _ := parser.SerializeJson(publicKeys)
		writer.WriteHeader(200)
		writer.Write(buffer)
		return
	}

	writer.WriteHeader(500)
	writer.Write([]byte("Failed"))
}

func (server *Server) handleServer(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	err := server.putterObj.Connect()
	defer server.putterObj.Disconnect()
	if err != nil {
		panic(err)
	}
	if request.Method == "POST" {
		managedServer := &manager.Server{}
		buffer, _ := ioutil.ReadAll(request.Body)
		err := parser.ParseYaml(managedServer, string(buffer))
		if err != nil {
			panic(err)
		}

		server.putterObj.AddServer(request.Context(), managedServer)

		writer.WriteHeader(200)
		writer.Write([]byte("OK"))
		return
	}

	writer.WriteHeader(500)
	writer.Write([]byte("Failed"))
}

func (server *Server) handleAccount(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	err := server.putterObj.Connect()
	defer server.putterObj.Disconnect()
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

		server.putterObj.AddAccount(request.Context(), account)

		writer.WriteHeader(200)
		writer.Write([]byte("OK"))
		return
	}

	writer.WriteHeader(500)
	writer.Write([]byte("Failed"))
}
