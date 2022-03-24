package core

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"

	"github.com/alikarimi999/shitcoin/database"
)

type DatabaseIterator struct {
	NextHash []byte
	DB       database.Database
}

func (c *Chain) NewIter() *DatabaseIterator {

	return &DatabaseIterator{c.LastBlock.BH.BlockHash, c.DB}

}

func (iter *DatabaseIterator) Next() *Block {

	block := ReadBlock(iter.DB, iter.NextHash)
	iter.NextHash = block.BH.PrevHash

	return block
}

func ReadBlock(d database.Database, hash []byte) *Block {

	b, err := d.DB.Get(hash, nil)
	if err != nil {
		log.Fatalln(err)
	}

	bl := Deserialize(b, new(Block))

	if block, ok := bl.(*Block); ok {
		return block

	}

	return nil
}

func saveBlockInDB(b Block, d *database.Database) error {

	block := Serialize(b)

	key := b.BH.BlockHash
	value := block

	err := d.DB.Put(key, value, nil)
	if err != nil {
		fmt.Println(err.Error())
	}

	err = d.DB.Put([]byte("last_hash"), b.BH.BlockHash, nil)
	fmt.Printf("Last Block in database is %x\n\n", b.BH.BlockHash)
	return err
}

func saveUTXOsInDB(u ChainState) error {

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
