package main

import (
	"flag"
	"fmt"

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
