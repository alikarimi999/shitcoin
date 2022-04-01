package core

import (
	"fmt"

	"github.com/alikarimi999/shitcoin/database"
)

func Loadchain(dbPath string, port int) *Chain {

	c, _ := NewChain(dbPath, port)
	block := ReadLastBlock(c.DB)

	if len(block.BH.BlockHash) == 0 {
		fmt.Println("There is no block in database")
		return c
	}

	c.LastBlock = *block
	c.ChainHeight = c.LastBlock.BH.BlockIndex + 1
	fmt.Printf("ChainHeight is %d\nlast block index: %d\n", c.ChainHeight, c.LastBlock.BH.BlockIndex)

	c.MemPool.Chainstate.Loadchainstate()

	return c

}

func ReadLastBlock(d database.Database) *Block {

	last_block := NewBlock()

	lh, err := d.DB.Get([]byte("last_hash"), nil)
	if err != nil {
		return last_block
	}

	b, err := d.DB.Get(lh, nil)
	if err != nil {
		return last_block
	}

	i := Deserialize(b, last_block)

	if block, ok := i.(*Block); ok {
		return block

	}

	return last_block
}
