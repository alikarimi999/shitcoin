package core

import (
	"fmt"
	"log"
	"time"

	"github.com/alikarimi999/shitcoin/database"
)

// Creat genesis Block
func (c *Chain) creatGenesis(to Address, amount float64) error {

	pkh := Add2PKH(to)
	genesis_block := &Block{
		BH: &blockHeader{
			Timestamp:  time.Now().UnixNano(),
			PrevHash:   []byte{},
			BlockIndex: 0,
			Difficulty: 10,
		},
		Transactions: []*Transaction{CoinbaseTx(pkh, 100)},
	}

	genesis_block.BH.BlockHash = genesis_block.Hash()

	c.Blocks = append(c.Blocks, genesis_block)
	c.LastBlock = genesis_block

	// we temprory add genesis output to mempool utxo set
	c.MemPool.Chainstate.UpdateUtxoSet(genesis_block.Transactions[0])

	err := SaveGenInDB(*genesis_block, &c.DB)
	if err != nil {
		log.Fatalln(err)
	}
	c.ChainHeight++
	fmt.Printf("Genesis Block added to database\n")
	c.SyncUtxoSet()
	return nil

}

func SaveGenInDB(b Block, d *database.Database) error {

	err := saveBlockInDB(b, d)
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
