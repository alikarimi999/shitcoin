package network

import (
	"bytes"
	"fmt"

	"github.com/alikarimi999/shitcoin/core/types"
)

// this function check validation of block that mined by another node
func BlockValidator(b types.Block, ch *types.ChainState, last_block types.Block) bool {
	ch.Mu.Lock()
	defer ch.Mu.Unlock()
	if b.BH.BlockIndex-1 == last_block.BH.BlockIndex && bytes.Equal(b.BH.PrevHash, last_block.BH.BlockHash) && b.Validate_hash() {

		if utxos, valid := ch.Validate_blk_trx(b); valid {
			fmt.Println(valid)
			ch.Utxos = utxos
			return true
		}
		return false
	}
	return false
}
