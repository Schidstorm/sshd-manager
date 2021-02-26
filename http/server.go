package http

import (
	"context"
	"encoding/json"
	"errors"
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
	http.Handle("/account", jsonOutputMiddleware(http.HandlerFunc(server.handleAccount)))
	http.Handle("/server", jsonOutputMiddleware(http.HandlerFunc(server.handleServer)))
	http.Handle("/serverKeys", jsonOutputMiddleware(http.HandlerFunc(server.handleServerKeys)))

	logrus.Infof("connecting to etcd")
	err := server.putterObj.Connect()
	defer server.putterObj.Disconnect()
	if err != nil {
		return err
	}

	logrus.Infof("listening on %s", addr)
	return http.ListenAndServe(addr, nil)
}

func (server *Server) handleServerKeys(writer http.ResponseWriter, request *http.Request) {
	response := request.Context().Value("response").(*Response)

	if request.Method == "GET" {
		serverKey := request.URL.Query().Get("key")
		publicKeys := server.putterObj.GetPublicKeysByServerKey(request.Context(), serverKey)
		response.Ok(publicKeys)
	} else {
		response.Err(errors.New("method not allowed"))
	}
}

func (server *Server) handleServer(writer http.ResponseWriter, request *http.Request) {
	response := request.Context().Value("response").(*Response)
	defer request.Body.Close()

	if request.Method == "POST" {
		managedServer := &manager.Server{}
		buffer, _ := ioutil.ReadAll(request.Body)
		err := parser.ParseYaml(managedServer, string(buffer))
		if err != nil {
			response.Err(err)
			return
		}

		server.putterObj.AddServer(request.Context(), managedServer)

		response.Ok(nil)
	} else if request.Method == "GET" {
		serverKey := request.URL.Query().Get("key")
		if serverKey == "" {
			response.Ok(server.putterObj.ListServers(request.Context()))
		} else {
			response.Ok(server.putterObj.GetServer(request.Context(), serverKey))
		}
	} else if request.Method == "DELETE" {
		serverKey := request.URL.Query().Get("key")
		server.putterObj.DeleteServer(request.Context(), serverKey)
		response.Ok(nil)
	} else {
		response.Err(errors.New("method not allowed"))
	}
}

func (server *Server) handleAccount(writer http.ResponseWriter, request *http.Request) {
	response := request.Context().Value("response").(*Response)
	defer request.Body.Close()

	if request.Method == "POST" {
		account := &manager.Account{}
		buffer, _ := ioutil.ReadAll(request.Body)
		err := parser.ParseYaml(account, string(buffer))
		if err != nil {
			response.Err(err)
			return
		}

		server.putterObj.AddAccount(request.Context(), account)
		response.Ok(nil)
	} else if request.Method == "GET" {
		accountKey := request.URL.Query().Get("key")
		if accountKey == "" {
			response.Ok(server.putterObj.ListAccounts(request.Context()))
		} else {
			response.Ok(server.putterObj.GetAccount(request.Context(), accountKey))
		}
	} else if request.Method == "DELETE" {
		accountKey := request.URL.Query().Get("key")
		server.putterObj.DeleteAccount(request.Context(), accountKey)
		response.Ok(nil)
	} else {
		response.Err(errors.New("method not allowed"))
	}
}

func jsonOutputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := &Response{}
		responseCtx := r.WithContext(context.WithValue(r.Context(), "response", response))
		next.ServeHTTP(w, responseCtx)
		writeResponse(w, response)
	})
}

func writeResponse(writer http.ResponseWriter, response *Response) {
	writer.Header().Set("Content-Type", "application/json")
	if response.Success {
		writer.WriteHeader(200)
	} else {
		writer.WriteHeader(500)
	}
	jsonData, err := json.Marshal(*response)
	if err != nil {
		logrus.Error(err)
		return
	}

	writer.Write(jsonData)
}
