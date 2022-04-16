package database

import (
	"time"

	"github.com/alikarimi999/shitcoin/core/types"
)

// tmplBlk is a block tmeplate that consits of block and some metadata that saved in database
type tmplBlk struct {
	Miner  string    // node that mined block for first time
	Sender string    // which node send block to this node for first time
	Time   time.Time // when block saved in database

	Block types.Block
}

type tmplState struct {
	Chain_height uint64
	Time         time.Time
	Owner        []byte
	Utxos        []*types.UTXO
}
