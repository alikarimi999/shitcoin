package core

import (
	"fmt"
	"log"

	"github.com/alikarimi999/shitcoin/core/types"
)

func (c *Chain) AddBlockInDB(b *types.Block) {

	// Saving valid block in database
	err := b.SaveBlockInDB(&c.DB)
	if err != nil {
		log.Fatalf("... Block %x did not add to database\n\n", b.BH.BlockHash)
	}
	fmt.Printf("... Block %d with hash %x successfully added to database\n\n", b.BH.BlockIndex, b.BH.BlockHash)

}
