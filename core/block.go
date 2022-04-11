package core

import (
	"log"
	"sync"

	"github.com/alikarimi999/shitcoin/core/types"
)

const (
	minerReward int = 15
)

func MinerReward(miner types.Address, amount int) *types.Transaction {
	pkh := types.Add2PKH(miner)
	reward := types.CoinbaseTx(pkh, amount)
	return reward
}

func (c *Chain) AddBlockInDB(b *types.Block, mu *sync.Mutex) {

	// Saving valid block in database
	err := b.SaveBlockInDB(&c.DB, mu)
	if err != nil {
		log.Fatalf("Block %x did not add to database\n\n", b.BH.BlockHash)
	}
	log.Printf("Block %d with hash %x successfully added to database\n\n", b.BH.BlockIndex, b.BH.BlockHash)

}
