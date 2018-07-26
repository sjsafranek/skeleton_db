package main

import (
	"flag"
	"fmt"

	"github.com/sjsafranek/gopass/lib"
	"github.com/sjsafranek/ligneous"
)

const DEFAULT_DB_FILE = "gonotes.db"

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
	// source: http://patorjk.com/software/taag/#p=display&f=Graceful&t=GOJOT2
	fmt.Println(`
  ___   __     __   __  ____  ____
 / __) /  \  _(  ) /  \(_  _)(___ \
( (_ \(  O )/ \) \(  O ) )(   / __/
 \___/ \__/ \____/ \__/ (__) (____)
    `)

	DB = lib.OpenDb(DB_FILE)
	defer DB.Close()
	RunTcpServer()
}
