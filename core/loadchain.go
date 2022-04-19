package core

import (
	"fmt"
	"log"
)

func Loadchain(dbPath string, port int, miner []byte) (*Chain, error) {

	c, err := NewChain(dbPath, port, miner)
	last_block, err := c.DB.LastBlock(nil)
	if err != nil {
		log.Println(err.Error())
		return c, nil
	}

	// TODO: use BlkIterator
	for i := 0; i <= int(last_block.BH.BlockIndex); i++ {
		b, err := c.DB.GetBlockI(uint64(i), nil)
		if err != nil {
			break
		}
		c.LastBlock = *b
		c.ChainHeight++
		c.Node.NodeHeight++
	}

	err = c.ChainState.Load()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("ChainHeight is %d\nlast block index: %d\n", c.ChainHeight, c.LastBlock.BH.BlockIndex)

	return c, err

}
