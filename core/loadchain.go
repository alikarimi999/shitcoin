package core

import (
	"errors"
	"fmt"
	"log"

	"github.com/alikarimi999/shitcoin/core/types"
	"github.com/alikarimi999/shitcoin/database"
)

func Loadchain(dbPath string, port int, miner []byte) (*Chain, error) {

	c, err := NewChain(dbPath, port, miner)
	block, err := ReadLastBlock(c.DB)
	if err != nil {
		log.Println(err.Error())
		return c, nil
	}

	c.LastBlock = *block
	c.ChainHeight = c.LastBlock.BH.BlockIndex + 1
	fmt.Printf("ChainHeight is %d\nlast block index: %d\n", c.ChainHeight, c.LastBlock.BH.BlockIndex)

	// TODO: load chainstate from database

	return c, err

}

func ReadLastBlock(d database.Database) (*types.Block, error) {

	last_block := types.NewBlock()

	lh, err := d.DB.Get([]byte("last_hash"), nil)
	if err != nil {
		return last_block, errors.New("didn't found last_hash")
	}

	b, err := d.DB.Get(lh, nil)
	if err != nil {
		return last_block, errors.New("didn't found last_block")
	}

	i := Deserialize(b, last_block)

	if block, ok := i.(*types.Block); ok {
		return block, errors.New("didn't found last_block")

	}

	return last_block, nil
}
