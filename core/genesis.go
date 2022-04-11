package core

import (
	"log"
	"sync"

	"github.com/alikarimi999/shitcoin/core/types"
	"github.com/alikarimi999/shitcoin/database"
)

// Creat genesis Block
func (c *Chain) creatGenesis() {

	pkh := types.Add2PKH(c.MinerAdd)
	tx := types.CoinbaseTx(pkh, 15)

	c.ChainState.StateTransition(tx.SnapShot(), false)
	c.ChainState.MinerIsStarting(true)

	c.Miner.Start([]*types.Transaction{tx}, c.TxPool.GetWaitGroup())
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
