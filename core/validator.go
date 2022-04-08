package core

import "github.com/alikarimi999/shitcoin/core/types"

type Validator interface {
	ValidateTX(tx *types.Transaction) bool

	// sned block snapshot for preventing data race
	ValidateBlk(b *types.Block) bool
}

type validator struct {
	c *Chain
}

func NewValidator(c *Chain) *validator {
	return &validator{c: c}
}

func (v *validator) ValidateTX(tx *types.Transaction) bool {
	account := types.Account(types.Pub2Address(tx.TxInputs[0].PublicKey, false))
	tokens := v.c.State.GetTokens(account)

	return tx.IsValid(tokens)
}

func (v *validator) ValidateBlk(b *types.Block) bool {
	return v.c.Engine.VerifyBlock(b, v.c.State.GetStableSet(), v.c.LastBlock)
}
