package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/dmichael/go-multicast/multicast"
)

const (
	defaultListenerMulticastAddress    = "239.0.0.0:4321"
	defaultBroadcasterMulticastAddress = "239.0.0.0:1234"
	base_reply                         = "\033[31mgtfs-cli>\033[0m "
	// base_reply = ""
)

var (
	listener_address    = defaultListenerMulticastAddress
	broadcaster_address = defaultBroadcasterMulticastAddress
)

type CommandListener struct {
	ListenerAddress    string
	BroadcasterAddress string
	tcp_port           int
}

func (self *CommandListener) Run() {
	go multicast.Listen(self.ListenerAddress, self.udpHandler)
	go self.tcpServer()
}

func (self *CommandListener) getTcpListener(port int) (net.Listener, error) {
	address := fmt.Sprintf("localhost:%v", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		Log.Error("Error listening:", err.Error())
		return listener, err
	}
	Log.Info("Listening on " + address)
	return listener, nil
}

func (self *CommandListener) tcpServer() {
	// try to find a port to listen on
	self.tcp_port = 9622
	listener := (func() net.Listener {
		for {
			listener, err := self.getTcpListener(self.tcp_port)
			if nil == err {
				return listener
			}
			self.tcp_port++
		}
	})()

	Log.Infof("Waiting for commands on %v", self.tcp_port)

	// Close the listener when the application closes.
	defer listener.Close()

	for {
		// Listen for an incoming connection.
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			continue
		}
		// Handle connections in a new goroutine.
		go self.tcpHandler(conn)
	}
}

func (self *CommandListener) sendBroadcast(message string) {
	conn, err := multicast.NewBroadcaster(self.BroadcasterAddress)
	if err != nil {
		Log.Error(err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte(message))
	if err != nil {
		Log.Error("Couldn't send response %v\n", err)
	}
}

func (self *CommandListener) getStats() string {
	app_info := make(map[string]interface{})
	app_info["start_time"] = START_TIME
	app_info["version"] = fmt.Sprintf("%v %v", PROJECT_NAME, Version())
	app_info["name"] = config.Name
	app_info["api_pusher"] = db4iotPusher.GetRuntimeInfo()

	database_info := make(map[string]interface{})
	database_info["max_threads"] = MAX_THREADS
	database_info["entities_per_thread"] = VEHICLES_PER_WORKER
	app_info["database"] = database_info

	queues := make(map[string]interface{})
	queues["feed"] = len(FeedWorkQueue)
	queues["database"] = len(DatabaseWorkQueue)
	app_info["queues"] = queues

	ports := make(map[string]interface{})
	ports["tcp_listener"] = self.tcp_port
	ports["udp_listener"] = self.ListenerAddress
	ports["udp_broadcast"] = self.BroadcasterAddress
	app_info["ports"] = ports

	gtfs_info := make(map[string]interface{})
	gtfs_info["schedule_update_time"] = SCHEDULE_UPDATE_TIME
	gtfs_info["schedule_source"] = config.Feed.GetSchedule()
	gtfs_info["vehicle_positions_source"] = config.Feed.GetVehiclePositions()
	gtfs_info["vehicle_positions_interval"] = VEHICLE_POSITIONS_INTERVAL
	gtfs_info["schedule_interval"] = SCHEDULE_INTERVAL

	Lock.RLock()
	entities_per_feed_5_minute := 0
	entities_per_feed_15_minute := 0
	entities_per_feed_60_minute := 0
	for k := range RtFeedStats {
		if time.Since(k) <= 5*time.Minute {
			entities_per_feed_5_minute += RtFeedStats[k]
		}
		if time.Since(k) <= 15*time.Minute {
			entities_per_feed_15_minute += RtFeedStats[k]
		}
		if time.Since(k) <= 60*time.Minute {
			entities_per_feed_60_minute += RtFeedStats[k]
		}
	}
	Lock.RUnlock()
	gtfs_rtfeed_info := make(map[string]interface{})
	gtfs_rtfeed_info["entities_per_feed_5_minute"] = entities_per_feed_5_minute
	gtfs_rtfeed_info["entities_per_feed_15_minute"] = entities_per_feed_15_minute
	gtfs_rtfeed_info["entities_per_feed_60_minute"] = entities_per_feed_60_minute
	gtfs_info["stats"] = gtfs_rtfeed_info

	app_info["gtfs"] = gtfs_info

	data, _ := json.Marshal(app_info)
	return string(data)
}

func (self *CommandListener) pause() string {
	db4iotPusher.Pause = true
	results := make(map[string]interface{})
	results["datasource_id"] = db4iotPusher.DatasourceId
	results["ingesting"] = false
	data, _ := json.Marshal(results)
	return string(data)
}

func (self *CommandListener) resume() string {
	db4iotPusher.Pause = false
	results := make(map[string]interface{})
	results["datasource_id"] = db4iotPusher.DatasourceId
	results["ingesting"] = true
	data, _ := json.Marshal(results)
	return string(data)
}

func (self *CommandListener) udpHandler(src *net.UDPAddr, n int, b []byte) {
	Log.Infof("%v bytes read from %s %s", n, src, string(b))

	command := self.getCommandFromMessage(b)
	message := self.HandleCommand(command)

	if "" != message {
		go self.sendBroadcast(message)
	}
}

func (self *CommandListener) getCommandFromMessage(buffer []byte) string {
	b := bytes.Trim(buffer, "\x00")
	command := fmt.Sprintf("%v", string(b))
	command = strings.TrimSpace(command)
	command = strings.Replace(command, "\n", "", -1)
	return command
}

func (self *CommandListener) sendTcpReply(conn net.Conn, message string) {
	conn.Write([]byte(message + "\n"))
	conn.Write([]byte(base_reply))
}

// Handles incoming requests.
func (self *CommandListener) tcpHandler(conn net.Conn) {
	defer conn.Close()
	conn.Write([]byte(base_reply))

	exitflag := false

	for {
		// Make a buffer to hold incoming data.
		buffer := make([]byte, 1024)
		// Read the incoming connection into the buffer.
		_, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error reading:", err.Error())
			break
		}

		command := self.getCommandFromMessage(buffer)

		switch {
		case strings.HasPrefix(command, "help"):
			self.sendTcpReply(conn, "Commands: help pause resume stats exit")

		case strings.HasPrefix(command, "pause"):
			msg := self.HandleCommand("pause")
			self.sendTcpReply(conn, msg)

		case strings.HasPrefix(command, "resume"):
			msg := self.HandleCommand("resume")
			self.sendTcpReply(conn, msg)

		case strings.HasPrefix(command, "stats"):
			msg := self.HandleCommand("stats")
			self.sendTcpReply(conn, msg)

		case strings.HasPrefix(command, "quit"):
			fallthrough
		case strings.HasPrefix(command, "bye"):
			fallthrough
		case strings.HasPrefix(command, "exit"):
			exitflag = true
			break

		case "" == command:
			conn.Write([]byte(base_reply))

		case "" != command:
			self.sendTcpReply(conn, "Unknown command")
		}

		if exitflag {
			break
		}

		// Send a response back to person contacting us.
		// conn.Write([]byte(base_reply))

	}
}

func (self *CommandListener) HandleCommand(command string) string {
	message := ""

	switch {
	case strings.HasPrefix(command, "datasource"):
		results := make(map[string]interface{})
		results["datasource_id"] = db4iotPusher.DatasourceId
		data, _ := json.Marshal(results)
		message = string(data)

	case strings.HasPrefix(command, "pause"):
		if len(command) == len("pause") {
			message = self.pause()
		} else if strings.Contains(command, db4iotPusher.DatasourceId) {
			message = self.pause()
		}

	case strings.HasPrefix(command, "resume"):
		if len(command) == len("resume") {
			message = self.resume()
		} else if strings.Contains(command, db4iotPusher.DatasourceId) {
			message = self.resume()
		}

	case strings.HasPrefix(command, "stats"):
		if len(command) == len("stats") {
			message = self.getStats()
		} else if strings.Contains(command, db4iotPusher.DatasourceId) {
			message = self.getStats()
		}

	}

	return message
}
