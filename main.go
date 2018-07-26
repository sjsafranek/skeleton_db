package main

import (
	// "errors"
	"flag"
	"fmt"
	// "os"
	// "strings"
	// "time"

	"github.com/sjsafranek/gopass/lib"

	"github.com/sjsafranek/ligneous"
)

const DEFAULT_DB_FILE = "gonotes.db"

var logger ligneous.Log
var DB_FILE string = DEFAULT_DB_FILE
var DB lib.Database

// func usage() {
// 	fmt.Printf("gojot2 0.0.1\n\n")
// 	fmt.Printf("Usage:\n\tgojot2 [options...] action key [action_args...]\n\n")
// 	fmt.Println(" * action:\tThe action to preform. Supported action(s): GET, SET, DEL")
// 	fmt.Println(" * action_args:\tVariadic arguments provided to the requested action. Different actions require different arguments")
// 	fmt.Println("\n")
// }

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

	// args := flag.Args()

	DB = lib.OpenDb(DB_FILE)
	defer DB.Close()

	// result, err := (func() (string, error) {
	//
	// 	switch strings.ToLower(args[0]) {
	//
	// 	case "get":
	// 		return DB.Get("store", args[1], args[2])
	//
	// 	case "set":
	// 		return "success", DB.Set("store", args[1], args[2], args[3])
	//
	// 	case "run":
	// 		RunTcpServer()
	// 		return "success", nil
	// 	// case "encrypt":
	// 	// 	return cryptic.Encrypt(args[2], args[1])
	// 	//
	// 	// case "decrypt":
	// 	// 	return cryptic.Decrypt(args[2], args[1])
	//
	// 	default:
	// 		return "", errors.New("Unknown command")
	// 	}
	//
	// })()

	RunTcpServer()

	// if nil != err {
	// 	logger.Error(err)
	// 	os.Exit(1)
	// }
	//
	// logger.Info(result)

	// time.Sleep(100 * time.Millisecond)

}
