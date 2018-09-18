package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	// "strconv"

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
	readline.PcItem("SETCLIENTENCRYPTION"),
)

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

func readResponse() string {
	status, err := bufio.NewReader(DB_CONN).ReadString('\n')
	if nil != err {
		log.Fatal(err)
	}
	// log.Println(status)
	return status
}

func sendQuery(query string) string {
	payload := fmt.Sprintf("%v\r\n", query)
	fmt.Fprintf(DB_CONN, payload)
	return readResponse()
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

	client := ApiClient{ClientEncryption: true}
	var namespace string
	var passphrase string
	// var client_encryption bool = true

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
		parts := strings.Split(line, " ")
		command := strings.ToLower(parts[0])

		// testing
		setPasswordCfg := l.GenPasswordConfig()
		setPasswordCfg.SetListener(func(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
			l.SetPrompt(fmt.Sprintf("Enter password(%v): ", len(line)))
			l.Refresh()
			return nil, 0, false
		})
		//.end

		switch {

		case strings.HasPrefix(command, "setclientencryption"):
			if 2 == len(parts) {
				boolean := parts[1]
				if "on" == strings.ToLower(boolean) {
					client.ClientEncryption = true
				} else if "off" == strings.ToLower(boolean) {
					client.ClientEncryption = false
				}
				continue
			}
			log.Println("Error! Incorrect usage")
			log.Println("SETCLIENTENCRYPTION <on||off>")

		case strings.HasPrefix(command, "setnamespace"):
			if 2 == len(parts) {
				namespace = parts[1]
				client.Namespace = namespace
				continue
			}
			log.Println("Error! Incorrect usage")
			log.Println("SETNAMESPACE <namespace>")

		case strings.HasPrefix(command, "setpassphrase"):
			pswd, err := l.ReadPasswordWithConfig(setPasswordCfg)
			if err == nil {
				passphrase = string(pswd)
			}

		case strings.HasPrefix(command, "getnamespace"):
			// log.Println(namespace)
			log.Println(client.Namespace)

		case strings.HasPrefix(command, "getpassphrase"):
			log.Println(passphrase)

		case strings.HasPrefix(command, "del"):
			var key string

			if 2 == len(parts) {
				if "del" == parts[0] {
					key = parts[1]
					query := fmt.Sprintf(`{"method":"del","data":{"key":"%v","namespace":"%v","passphrase":"%v"}}`, key, namespace, passphrase)
					log.Println(sendQuery(query))
					continue
				}
			}
			log.Println("Error! Incorrect usage")
			log.Println("DEL <key>")

		case strings.HasPrefix(command, "get"):
			var key string

			if 2 == len(parts) {
				if "get" == parts[0] {
					key = parts[1]
					// query := fmt.Sprintf(`{"method":"get","data":{"key":"%v","namespace":"%v","passphrase":"%v"}}`, key, namespace, passphrase)
					// log.Println(sendQuery(query))
					// results := sendQuery(query)
					value, err := client.Get(key, passphrase)
					if nil != err {
						log.Println(err)
						continue
					}
					log.Println(value)
					continue
				}
			}
			log.Println("Error! Incorrect usage")
			log.Println("GET <key>")

		case strings.HasPrefix(command, "set"):
			var key string
			var value string

			if "set" == parts[0] {
				key = parts[1]

				i1 := strings.Index(line, "'")
				i2 := strings.LastIndex(line, "'")
				value = line[i1+1 : i2]

				// query := fmt.Sprintf(`{"method":"set","data":{"key":"%v","value":"%v","namespace":"%v","passphrase":"%v"}}`, key, value, namespace, passphrase)
				// log.Println(sendQuery(query))
				status, err := client.Set(key, value, passphrase)
				if nil != err {
					log.Println(err)
					continue
				}
				log.Println(status)

				continue
			}

			log.Println("Error! Incorrect usage")
			log.Println("SET <key> <value>")

		case command == "help":
			usage(l.Stderr())

		case strings.HasPrefix(command, "keys"):
			query := fmt.Sprintf(`{"method": "keys", "data":{"namespace":"%v"}}`, namespace)
			log.Println(sendQuery(query))

		case strings.HasPrefix(command, "namespaces"):
			query := `{"method": "namespaces"}`
			log.Println(sendQuery(query))

		case command == "bye":
			goto exit

		case command == "exit":
			goto exit

		case command == "quit":
			goto exit

		case line == "":
		default:
			// log.Println("you said:", strconv.Quote(line))
		}
	}
exit:
}
