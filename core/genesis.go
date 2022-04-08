package core

import (
	"log"
	"sync"

	"github.com/alikarimi999/shitcoin/core/types"
	"github.com/alikarimi999/shitcoin/database"
)

// Creat genesis Block
func (c *Chain) creatGenesis() {
	c.Miner.Start([]*types.Transaction{})
}

func SaveGenInDB(b types.Block, d *database.Database) error {

	err := b.SaveBlockInDB(d, &sync.Mutex{})
	if err != nil {
		log.Fatalln(err)
	}

	key := []byte("genesis_block")
	value := b.BH.BlockHash

	err = d.DB.Put(key, value, nil)
	if err != nil {
		log.Fatalln(err)
	}

	return nil

}
