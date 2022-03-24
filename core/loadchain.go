package core

import (
	"fmt"

	"github.com/alikarimi999/shitcoin/database"
)

func Loadchain(dbPath string) *Chain {

	c, _ := NewChain(dbPath)
	c.LastBlock = ReadLastBlock(c.DB)

	if c.LastBlock == nil {
		fmt.Println("There is no block in database")
		return c
	}
	c.ChainHeight = c.LastBlock.BH.BlockIndex + 1
	fmt.Printf("ChainHeight is %d\nlast block index: %d\n", c.ChainHeight, c.LastBlock.BH.BlockIndex)

	c.MemPool.Chainstate.Loadchainstate()

	return c

}

func ReadLastBlock(d database.Database) *Block {

	lh, err := d.DB.Get([]byte("last_hash"), nil)
	if err != nil {
		return nil
	}

	b, err := d.DB.Get(lh, nil)
	if err != nil {
		return nil
	}

	bl := Deserialize(b, new(Block))

	if block, ok := bl.(*Block); ok {
		return block

	}

	return nil
}
