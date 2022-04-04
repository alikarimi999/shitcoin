package consensus

import (
	"github.com/alikarimi999/shitcoin/core/types"
)

type Engin interface {

	// VerifyBlock checks whether a header conforms to the consensus rules of a
	// given engine.
	VerifyBlock(ch *types.ChainState, last_block types.Block) bool

	// IsRunning checks whether consensus engine is searching for nonce or not
	IsRunning() bool

	// Start start searching for  nonce base on consensus alghorithm of given engine
	Start(b *types.Block) bool
	// Pause engine
	Pause()
	// Close engine
	Close()
	// Resume engine
	Resume()
	GetHash() []byte
}
