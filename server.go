package main

import (
	"encoding/json"
	"net"
	"runtime"
	"time"

	"github.com/sjsafranek/socket2em"
)

const (
	// DEFAULT_PORT default tcp port
	DEFAULT_PORT = 9622
	// DEFAULT_NAMESPACE default bucket name for database
	DEFAULT_NAMESPACE = "store"
	//
	DEFAULT_HOST = "0.0.0.0"
)

var (
	// TCP_SERVER tcp server
	TCP_SERVER socket2em.Server
	// PORT port to listen to
	PORT = DEFAULT_PORT
	//
	HOST = DEFAULT_HOST
)

func init() {
	TCP_SERVER = NewTcpServer()
}

func parseRawJsonMessage(raw json.RawMessage) map[string]string {
	var data map[string]string
	json.Unmarshal(raw, &data)
	logger.Info(data)
	return data
}

// NewTcpServer creates and returns socket2em.Server tcp server
func NewTcpServer() socket2em.Server {

	startTime := time.Now()

	server := socket2em.Server{
		LoggingHandler: func(message string) { logger.Info(message) },
		Port:           PORT,
		Host:           HOST,
	}

	// Simple ping method
	server.RegisterMethod("ping", func(message socket2em.Message, conn net.Conn) {
		// {"method": "ping"}
		server.HandleSuccess(`{"message": "pong"}`, conn)
	})

	// Returns runtime and system information
	server.RegisterMethod("get_runtime_stats", func(message socket2em.Message, conn net.Conn) {
		// {"method": "get_runtime_stats"}
		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)
		results := make(map[string]interface{})
		results["NumGoroutine"] = runtime.NumGoroutine()
		results["Alloc"] = ms.Alloc / 1024
		results["TotalAlloc"] = ms.TotalAlloc / 1024
		results["Sys"] = ms.Sys / 1024
		results["NumGC"] = ms.NumGC
		results["Registered"] = startTime.UTC()
		results["Uptime"] = time.Since(startTime).Seconds()
		results["NumCPU"] = runtime.NumCPU()
		results["GOOS"] = runtime.GOOS
		server.SendResponseFromStruct(results, conn)
	})

	// Returns runtime and system information
	server.RegisterMethod("num_clients", func(message socket2em.Message, conn net.Conn) {
		// {"method": "num_clients"}
		results := make(map[string]interface{})
		results["num_clients"] = TCP_SERVER.GetNumClients()
		server.SendResponseFromStruct(results, conn)
	})

	server.RegisterMethod("namespaces", func(message socket2em.Message, conn net.Conn) {
		// {"method": "namespaces"}
		tables, err := DB.Tables()
		if nil != err {
			logger.Error(err)
			TCP_SERVER.HandleError(err, conn)
			return
		}
		results := make(map[string]interface{})
		results["namespaces"] = tables
		server.SendResponseFromStruct(results, conn)
	})

	// Get keys
	server.RegisterMethod("keys", func(message socket2em.Message, conn net.Conn) {
		// {"method": "keys", "data":{"namespace":"test"}}
		data := parseRawJsonMessage(message.Data)
		namespace := DEFAULT_NAMESPACE
		if "" != data["namespace"] {
			namespace = data["namespace"]
		}

		keys, err := DB.Keys(namespace)
		if nil != err {
			logger.Error(err)
			TCP_SERVER.HandleError(err, conn)
			return
		}
		results := make(map[string]interface{})
		results["keys"] = keys
		server.SendResponseFromStruct(results, conn)
	})

	// Set value
	server.RegisterMethod("set", func(message socket2em.Message, conn net.Conn) {
		// {"method": "set", "data":{"key":"stefan","value":"rocks","passphrase":"test"}}
		data := parseRawJsonMessage(message.Data)
		namespace := DEFAULT_NAMESPACE
		if "" != data["namespace"] {
			namespace = data["namespace"]
		}

		err := DB.CreateTable(namespace)
		if nil != err {
			logger.Error(err)
			server.HandleError(err, conn)
			return
		}

		results := make(map[string]interface{})
		err = Set(namespace, data["key"], data["value"], data["passphrase"])
		if nil != err {
			logger.Error(err)
			server.HandleError(err, conn)
			return
		}

		server.SendResponseFromStruct(results, conn)
	})

	// Get key value
	server.RegisterMethod("get", func(message socket2em.Message, conn net.Conn) {
		// {"method": "get", "data":{"key":"stefan","passphrase":"test"}}
		data := parseRawJsonMessage(message.Data)
		namespace := DEFAULT_NAMESPACE
		if "" != data["namespace"] {
			namespace = data["namespace"]
		}

		results := make(map[string]interface{})
		val, err := Get(namespace, data["key"], data["passphrase"])
		if nil != err {
			logger.Error(err)
			server.HandleError(err, conn)
			return
		}

		results["value"] = val
		server.SendResponseFromStruct(results, conn)
	})

	// Delete key
	server.RegisterMethod("del", func(message socket2em.Message, conn net.Conn) {
		// {"method": "del", "data":{"key":"stefan","passphrase":"test"}}
		data := parseRawJsonMessage(message.Data)
		namespace := DEFAULT_NAMESPACE
		if "" != data["namespace"] {
			namespace = data["namespace"]
		}

		results := make(map[string]interface{})
		err := DB.Remove(namespace, data["key"], data["passphrase"])
		if nil != err {
			logger.Error(err)
			server.HandleError(err, conn)
			return
		}

		server.SendResponseFromStruct(results, conn)
	})

	return server

}
