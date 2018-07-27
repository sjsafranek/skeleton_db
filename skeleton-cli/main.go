package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	// "io/ioutil"
	"log"
	// "strconv"
	"strings"
	// "time"
	"net"

	"github.com/chzyer/readline"
)

const (
	DEFAULT_DATABASE_SERVER_ADDRESS = "localhost:9622"
)

var (
	DATABASE_SERVER_ADDRESS = DEFAULT_DATABASE_SERVER_ADDRESS
	DB_CONN                 net.Conn
)

func init() {
	// get command line args
	flag.StringVar(&DATABASE_SERVER_ADDRESS, "a", DEFAULT_DATABASE_SERVER_ADDRESS, "database server address")
	flag.Parse()
}

func usage(w io.Writer) {
	io.WriteString(w, "commands:\n")
	io.WriteString(w, completer.Tree("    "))
}

var completer = readline.NewPrefixCompleter(
	// readline.PcItem("mode",
	// 	readline.PcItem("vi"),
	// 	readline.PcItem("emacs"),
	// ),
	// readline.PcItem("login"),
	// readline.PcItem("say",
	// 	readline.PcItemDynamic(listFiles("./"),
	// 		readline.PcItem("with",
	// 			readline.PcItem("following"),
	// 			readline.PcItem("items"),
	// 		),
	// 	),
	readline.PcItem("KEYS"),
	readline.PcItem("NAMESPACES"),
	readline.PcItem("SET"),
	readline.PcItem("GET"),
	readline.PcItem("DEL"),
	readline.PcItem("BYE"),
	readline.PcItem("EXIT"),
	readline.PcItem("HELP"),
	readline.PcItem("SETNAMESPACE"),
	readline.PcItem("SETPASSPHRASE"),
	readline.PcItem("GETNAMESPACE"),
	readline.PcItem("GETPASSPHRASE"),
)

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

func readResponse() {
	status, err := bufio.NewReader(DB_CONN).ReadString('\n')
	if nil != err {
		log.Fatal(err)
	}
	log.Println(status)
}

func sendQuery(query string) {
	payload := fmt.Sprintf("%v\r\n", query)
	fmt.Fprintf(DB_CONN, payload)
	readResponse()
}

func main() {
	conn, err := net.Dial("tcp", DATABASE_SERVER_ADDRESS)
	if nil != err {
		log.Fatal(err)
	}

	DB_CONN = conn
	defer DB_CONN.Close()

	l, err := readline.NewEx(&readline.Config{
		Prompt:              "\033[31m[skeleton]#\033[0m ",
		HistoryFile:         "history.skeleton",
		AutoComplete:        completer,
		InterruptPrompt:     "^C",
		EOFPrompt:           "exit",
		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		panic(err)
	}
	defer l.Close()

	var namespace string
	var passphrase string

	log.SetOutput(l.Stderr())
	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		line = strings.ToLower(line)
		parts := strings.Split(line, " ")
		// log.Println(parts)

		switch {

		case strings.HasPrefix(line, "setnamespace"):
			if 2 == len(parts) {
				namespace = parts[1]
				continue
			}
			log.Println("Error! Incorrect usage")
			log.Println("SETNAMESPACE <namespace>")

		case strings.HasPrefix(line, "setpassphrase"):
			if 2 == len(parts) {
				passphrase = parts[1]
				continue
			}
			log.Println("Error! Incorrect usage")
			log.Println("SETPASSPHRASE <passphrase>")

		case strings.HasPrefix(line, "getnamespace"):
			log.Println(namespace)

		case strings.HasPrefix(line, "getpassphrase"):
			log.Println(passphrase)

		case strings.HasPrefix(line, "del"):
			var key string

			if 2 == len(parts) {
				if "del" == parts[0] {
					key = parts[1]
					query := fmt.Sprintf(`{"method":"del","data":{"key":"%v","namespace":"%v","passphrase":"%v"}}`, key, namespace, passphrase)
					sendQuery(query)
					continue
				}
			}
			log.Println("Error! Incorrect usage")
			log.Println("DEL <key>")

		case strings.HasPrefix(line, "get"):
			var key string

			if 2 == len(parts) {
				if "get" == parts[0] {
					key = parts[1]
					query := fmt.Sprintf(`{"method":"get","data":{"key":"%v","namespace":"%v","passphrase":"%v"}}`, key, namespace, passphrase)
					sendQuery(query)
					continue
				}
			}
			log.Println("Error! Incorrect usage")
			log.Println("GET <key>")

		case strings.HasPrefix(line, "set"):
			var key string
			var value string

			if 3 == len(parts) {
				if "set" == parts[0] {
					key = parts[1]
					value = parts[2]
					query := fmt.Sprintf(`{"method":"set","data":{"key":"%v","value":"%v","namespace":"%v","passphrase":"%v"}}`, key, value, namespace, passphrase)
					sendQuery(query)
					continue
				}
			}
			log.Println("Error! Incorrect usage")
			log.Println("SET <key> <value>")

		case line == "help":
			usage(l.Stderr())

		case strings.HasPrefix(line, "keys"):
			query := fmt.Sprintf(`{"method": "keys", "data":{"namespace":"%v"}}`, namespace)
			sendQuery(query)

		case strings.HasPrefix(line, "namespaces"):
			query := `{"method": "namespaces"}`
			sendQuery(query)

		case line == "bye":
			goto exit

		case line == "":
		default:
			// log.Println("you said:", strconv.Quote(line))
		}
	}
exit:
}
