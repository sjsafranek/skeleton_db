package main

import (
	"github.com/sjsafranek/DiffStore"
)

func Get(namespace, key, passphrase string) (diffstore.DiffStore, error) {
	var ddata diffstore.DiffStore
	data, err := DB.Get(namespace, key, passphrase)
	if nil != err {
		return ddata, err
	}
	ddata.Decode([]byte(data))
	return ddata, err
}

func Set(namespace, key, value, passphrase string) error {
	ddata, err := Get(namespace, key, passphrase)
	if nil != err {
		if err.Error() == "Not found" {
			// create new diffstore if key not found in database
			ddata = diffstore.New()
		} else {
			return err
		}
	}
	ddata.Update(value)

	enc, err := ddata.Encode()
	if nil != err {
		return err
	}

	return DB.Set(namespace, key, string(enc), passphrase)
}

//
// func Remove(key, passphrase string) error {
// 	return DB.Remove("store", key, passphrase)
// }
