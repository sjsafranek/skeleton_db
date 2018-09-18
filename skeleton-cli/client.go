package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"

	"github.com/sjsafranek/goutils/cryptic"
)

type ApiClient struct {
	Namespace        string
	ClientEncryption bool
	Conn             net.Conn
}

func (self *ApiClient) Disconnect() {
	if nil != self.Conn {
		self.Conn.Close()
	}
}

func (self *ApiClient) Connect(address string) error {
	self.Disconnect()

	conn, err := net.Dial("tcp", address)
	if nil != err {
		return err
	}

	self.Conn = conn
	return err
}

func (self *ApiClient) recieve() (string, error) {
	return bufio.NewReader(self.Conn).ReadString('\n')
}

func (self *ApiClient) send(query string) {
	payload := fmt.Sprintf("%v\r\n", query)
	fmt.Fprintf(self.Conn, payload)
}

func (self *ApiClient) sendAndReceive(query string) (ApiResponse, error) {
	if nil == self.Conn {
		return ApiResponse{}, errors.New("Not connected to database server")
	}

	self.send(query)
	results, err := self.recieve()
	if nil != err {
		return ApiResponse{}, err
	}
	return self.parseResponse(results)
}

func (self *ApiClient) parseResponse(results string) (ApiResponse, error) {
	var response ApiResponse
	err := json.Unmarshal([]byte(results), &response)

	if nil != err {
		return response, err
	}

	if "error" == response.Status {
		return response, errors.New(response.Error)
	}

	return response, err
}

func (self *ApiClient) Get(key, passphrase string) (string, error) {
	var query string
	if self.ClientEncryption {
		query = fmt.Sprintf(`{"method":"get","data":{"key":"%v","namespace":"%v","passphrase":"%v"}}`, key, self.Namespace, "")
	} else {
		query = fmt.Sprintf(`{"method":"get","data":{"key":"%v","namespace":"%v","passphrase":"%v"}}`, key, self.Namespace, passphrase)
	}

	response, err := self.sendAndReceive(query)
	if nil != err {
		return "", err
	}

	if self.ClientEncryption {
		garbage := response.Data.Value.Value
		value, err := cryptic.Decrypt(passphrase, garbage)
		return value, err
	}
	return response.Data.Value.Value, nil
}

func (self *ApiClient) Set(key, value, passphrase string) (string, error) {
	if self.ClientEncryption {
		garbage, _ := cryptic.Encrypt(passphrase, value)
		value = garbage
		passphrase = ""
	}

	query := fmt.Sprintf(`{"method":"set","data":{"key":"%v","value":"%v","namespace":"%v","passphrase":"%v"}}`, key, value, self.Namespace, passphrase)
	response, err := self.sendAndReceive(query)
	if nil != err {
		return "", err
	}

	return response.Status, nil
}

func (self *ApiClient) Del(key, passphrase string) (string, error) {
	query := fmt.Sprintf(`{"method":"del","data":{"key":"%v","namespace":"%v","passphrase":"%v"}}`, key, self.Namespace, passphrase)

	response, err := self.sendAndReceive(query)
	if nil != err {
		return "", err
	}
	return response.Status, nil
}

func (self *ApiClient) Keys() ([]string, error) {
	query := fmt.Sprintf(`{"method": "keys", "data":{"namespace":"%v"}}`, self.Namespace)

	response, err := self.sendAndReceive(query)
	if nil != err {
		return []string{}, err
	}

	return response.Data.Keys, nil
}

func (self *ApiClient) Namespaces() ([]string, error) {
	query := `{"method": "namespaces"}`

	response, err := self.sendAndReceive(query)
	if nil != err {
		return []string{}, err
	}
	return response.Data.Namespaces, nil
}
