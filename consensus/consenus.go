package consensus

import "github.com/alikarimi999/shitcoin/core"

type Engin interface {

	// VerifyBlock checks whether a header conforms to the consensus rules of a
	// given engine.
	VerifyBlock(ch *core.ChainState, last_block core.Block) bool

	// IsRunning checks whether consensus engine is searching for nonce or not
	IsRunning() bool

	// Start start searching for  nonce base on consensus alghorithm of given engine
	Start() ([]byte, error)
	// Pause engine
	Pause()
	// Abort engine
	Abort()
	// Resume engine
	Resume()
}
