package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sjsafranek/gopass/lib"
	"github.com/sjsafranek/ligneous"
)

const DEFAULT_DB_FILE = "skeleton.db"

var logger ligneous.Log
var DB_FILE string = DEFAULT_DB_FILE
var DB lib.Database

func init() {
	// initialize logger
	logger = ligneous.NewLogger()

	// get command line args
	flag.StringVar(&DB_FILE, "db", DEFAULT_DB_FILE, "database file")
	flag.Parse()

	signal_queue := make(chan os.Signal)
	signal.Notify(signal_queue, syscall.SIGTERM)
	signal.Notify(signal_queue, syscall.SIGINT)
	go func() {
		sig := <-signal_queue
		logger.Debugf("caught sig: %+v", sig)
		logger.Info("Gracefully shutting down...")
		c := 10
		for {
			if 0 == TCP_SERVER.GetNumClients() || c == 0 {
				break
			}
			logger.Debug("Waiting for clients to close")
			TCP_SERVER.Broadcast(fmt.Sprintf("server is shutting down in %v seconds...", c))
			time.Sleep(1 * time.Second)
			c--
		}
		logger.Debug("Closing tcp clients...")
		TCP_SERVER.Shutdown()
		logger.Debug("Closing database connection...")
		DB.Close()
		logger.Debug("Shutting down...")
		time.Sleep(1 * time.Second)
		os.Exit(0)
	}()
}

func main() {
	// source: http://patorjk.com/software/taag/#p=display&f=Slant&t=Skeleton
	fmt.Println(`
   _____ __        __     __
  / ___// /_____  / /__  / /_____  ____
  \__ \/ //_/ _ \/ / _ \/ __/ __ \/ __ \
 ___/ / ,< /  __/ /  __/ /_/ /_/ / / / /
/____/_/|_|\___/_/\___/\__/\____/_/ /_/
`)

	DB = lib.OpenDb(DB_FILE)
	defer DB.Close()
	RunTcpServer()
}
