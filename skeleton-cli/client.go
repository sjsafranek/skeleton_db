package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/sjsafranek/goutils/cryptic"
)

type ApiClient struct {
	Namespace string
	// Passphrase       string
	ClientEncryption bool
}

func (self *ApiClient) Get(key, passphrase string) (string, error) {
	var query string
	if self.ClientEncryption {
		query = fmt.Sprintf(`{"method":"get","data":{"key":"%v","namespace":"%v","passphrase":"%v"}}`, key, self.Namespace, "")
	} else {
		query = fmt.Sprintf(`{"method":"get","data":{"key":"%v","namespace":"%v","passphrase":"%v"}}`, key, self.Namespace, passphrase)
	}
	results := sendQuery(query)
	response, err := self.parseResponse(results)
	if nil != err {
		return "", err
	}
	if "error" == response.Status {
		return "", errors.New(response.Error)
	}
	if self.ClientEncryption {
		garbage := response.Data.Value.Value
		value, err := cryptic.Decrypt(passphrase, garbage)
		return value, err
	}
	return response.Data.Value.Value, nil
}

func (self *ApiClient) parseResponse(results string) (ApiResponse, error) {
	var response ApiResponse
	err := json.Unmarshal([]byte(results), &response)
	return response, err
}

func (self *ApiClient) Set(key, value, passphrase string) (string, error) {
	if self.ClientEncryption {
		garbage, _ := cryptic.Encrypt(passphrase, value)
		value = garbage
		passphrase = ""
	}
	query := fmt.Sprintf(`{"method":"set","data":{"key":"%v","value":"%v","namespace":"%v","passphrase":"%v"}}`, key, value, self.Namespace, passphrase)
	results := sendQuery(query)
	response, err := self.parseResponse(results)
	if nil != err {
		return "", err
	}
	if "error" == response.Status {
		return "", errors.New(response.Error)
	}
	return response.Status, nil
}

//
// func (self *ApiClient) Del(key, value, passphrase string) {
//
// }
