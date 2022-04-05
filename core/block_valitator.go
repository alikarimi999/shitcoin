package core

import (
	"errors"

	"github.com/alikarimi999/shitcoin/consensus"
	"github.com/alikarimi999/shitcoin/core/types"
)

type BlockValidator struct {
	c      *Chain
	engine consensus.Engin
	cs     *types.ChainState
}

func NewBlockValidator(c *Chain, engine consensus.Engin) *BlockValidator {
	return &BlockValidator{
		c:      c,
		engine: engine,
		cs:     types.NewChainState(),
	}
}

// Validate Check if recieved block is valid or not
// db must set true if we want check block with chainstate that saved in database
func (bv *BlockValidator) ValidateBlock(b *types.Block, db bool) error {

	if db {
		bv.c.Chainstate.Loadchainstate()
		bv.cs = bv.c.Chainstate
	} else {
		bv.cs = bv.c.MemPool.Chainstate.SnapShot()
	}
	if bv.engine.VerifyBlock(b, bv.cs, bv.c.LastBlock) {
		return nil
	}
	return errors.New("block is not valid")

}

func (bv *BlockValidator) GetChainState() *types.ChainState {
	return bv.cs
}

func (bv *BlockValidator) Reset() {
	bv.cs.Clean()
}
