package database

import (
	"bytes"
	"encoding/gob"
	"log"
)

func Serialize(t interface{}) []byte {
	buff := bytes.Buffer{}

	encoder := gob.NewEncoder(&buff)
	err := encoder.Encode(t)
	if err != nil {
		log.Printf("database: %s\n", err.Error())
		return nil
	}
	return buff.Bytes()
}

func Deserialize(b []byte, t interface{}) (interface{}, error) {

	decoder := gob.NewDecoder(bytes.NewBuffer(b))

	err := decoder.Decode(t)
	if err != nil {
		log.Printf("database: %s\n", err.Error())
		return nil, err
	}

	return t, nil
}
