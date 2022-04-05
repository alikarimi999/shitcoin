package core

import "github.com/alikarimi999/shitcoin/core/types"

type Validator interface {
	ValidateBlock(b *types.Block, db bool) error
	GetChainState() *types.ChainState
	Reset()
}
