package main

import (
	"encoding/json"
	"net"
	"runtime"
	"time"

	"github.com/sjsafranek/socket2em"
)

var (
	TCP_SERVER socket2em.Server
	tcp_port   int = 9622
	startTime  time.Time
)

const (
	DEFAULT_NAMESPACE = "store"
)

func init() {
	startTime = time.Now()
}

func parseRawJsonMessage(raw json.RawMessage) map[string]string {
	var data map[string]string
	json.Unmarshal(raw, &data)
	logger.Info(data)
	return data
}

func RunTcpServer() {

	TCP_SERVER = socket2em.Server{
		LoggingHandler: func(message string) { logger.Info(message) },
		Port:           tcp_port,
	}

	// Simple ping method
	TCP_SERVER.RegisterMethod("ping", func(message socket2em.Message, conn net.Conn) {
		// {"method": "ping"}
		TCP_SERVER.HandleSuccess(`{"message": "pong"}`, conn)
	})

	// Returns runtime and system information
	TCP_SERVER.RegisterMethod("get_runtime_stats", func(message socket2em.Message, conn net.Conn) {
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
		TCP_SERVER.SendResponseFromStruct(results, conn)
	})

	// Returns runtime and system information
	TCP_SERVER.RegisterMethod("num_clients", func(message socket2em.Message, conn net.Conn) {
		// {"method": "num_clients"}
		results := make(map[string]interface{})
		results["num_clients"] = TCP_SERVER.GetNumClients()
		TCP_SERVER.SendResponseFromStruct(results, conn)
	})

	TCP_SERVER.RegisterMethod("namespaces", func(message socket2em.Message, conn net.Conn) {
		// {"method": "namespaces"}
		tables, err := DB.Tables()
		if nil != err {
			logger.Error(err)
			TCP_SERVER.HandleError(err, conn)
			return
		}
		results := make(map[string]interface{})
		results["namespaces"] = tables
		TCP_SERVER.SendResponseFromStruct(results, conn)
	})

	// Get keys
	TCP_SERVER.RegisterMethod("keys", func(message socket2em.Message, conn net.Conn) {
		// {"method": "keys"}
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
		TCP_SERVER.SendResponseFromStruct(results, conn)
	})

	// Set value
	TCP_SERVER.RegisterMethod("set", func(message socket2em.Message, conn net.Conn) {
		// {"method": "set", "data":{"key":"stefan","value":"rocks","passphrase":"test"}}
		data := parseRawJsonMessage(message.Data)
		namespace := DEFAULT_NAMESPACE
		if "" != data["namespace"] {
			namespace = data["namespace"]
		}

		err := DB.CreateTable(namespace)
		if nil != err {
			logger.Error(err)
			TCP_SERVER.HandleError(err, conn)
			return
		}

		results := make(map[string]interface{})
		err = Set(namespace, data["key"], data["value"], data["passphrase"])
		if nil != err {
			logger.Error(err)
			TCP_SERVER.HandleError(err, conn)
			return
		}

		TCP_SERVER.SendResponseFromStruct(results, conn)
	})

	// Get key value
	TCP_SERVER.RegisterMethod("get", func(message socket2em.Message, conn net.Conn) {
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
			TCP_SERVER.HandleError(err, conn)
			return
		}

		results["value"] = val
		TCP_SERVER.SendResponseFromStruct(results, conn)
	})

	// Delete key
	TCP_SERVER.RegisterMethod("del", func(message socket2em.Message, conn net.Conn) {
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
			TCP_SERVER.HandleError(err, conn)
			return
		}

		TCP_SERVER.SendResponseFromStruct(results, conn)
	})

	TCP_SERVER.Start()

}
