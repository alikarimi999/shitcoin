package core

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"

	"github.com/alikarimi999/shitcoin/core/types"
	"github.com/alikarimi999/shitcoin/database"
)

type DatabaseIterator struct {
	NextHash []byte
	DB       database.Database
}

func (c *Chain) NewIter() *DatabaseIterator {

	return &DatabaseIterator{c.LastBlock.BH.BlockHash, c.DB}

}

func (iter *DatabaseIterator) Next() *types.Block {

	block := ReadBlock(iter.DB, iter.NextHash)
	iter.NextHash = block.BH.PrevHash

	return block
}

func ReadBlock(d database.Database, hash []byte) *types.Block {

	b, err := d.DB.Get(hash, nil)
	if err != nil {
		log.Fatalln(err)
	}

	bl := Deserialize(b, new(types.Block))

	if block, ok := bl.(*types.Block); ok {
		return block

	}

	return nil
}

func saveUTXOsInDB(u types.ChainState) error {

	for account, utxos := range u.Utxos {
		key := []byte(account)
		value := Serialize(utxos)

		err := u.DB.DB.Put(key, value, nil)

		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("All Tokens for %s saved in database\n\n", account)

	}
	return nil

}

func Serialize(t interface{}) []byte {
	buff := bytes.Buffer{}

	encoder := gob.NewEncoder(&buff)
	encoder.Encode(t)

	return buff.Bytes()
}

func Deserialize(b []byte, t interface{}) interface{} {

	decoder := gob.NewDecoder(bytes.NewBuffer(b))

	err := decoder.Decode(t)

	if err != nil {
		log.Fatalln(err)
	}

	return t
}
