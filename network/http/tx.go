package network

import (
	"fmt"
	"log"

	"github.com/alikarimi999/shitcoin/core"
	"github.com/alikarimi999/shitcoin/core/types"
	"github.com/labstack/echo/v4"
)

func (o *Objects) getTrx(ctx echo.Context) error {
	var t types.Transaction
	err := ctx.Bind(&t)

	if err != nil {
		log.Panic(err)
	}

	log.Printf("Transaction %x recieved\n", t.TxID)

	if o.Ch.Validator.ValidateTX(&t) {
		o.Ch.State.StateTransition(t.SnapShot(), false)
		o.Ch.TxPool.UpdatePool(t.SnapShot(), false)
		log.Printf("Transaction %x is valid\n", t.TxID)

		// Broadcast transaction
		BroadTrx(o.Ch, t)

		ctx.String(200, fmt.Sprintf("Transaction added to MemPool\n"))
		return nil

	}
	log.Printf("Transaction %x is not valid\n", t.TxID)
	ctx.String(403, "")
	return nil
}

func (o *Objects) sendUtxoset(ctx echo.Context) error {

	account := ctx.QueryParam("account")
	msg := sendUtxoset(o.Ch, types.Account(account))

	ctx.JSONPretty(200, msg, "  ")
	return nil
}

func sendUtxoset(c *core.Chain, a types.Account) msgUTXOSet {

	var s msgUTXOSet
	s.Account = a
	utxos := c.State.GetTokens(a)

	for _, utxo := range utxos {
		// su := core.UTXO{utxo.Txid, utxo.Index, utxo.Txout}
		s.Utxos = append(s.Utxos, *utxo)
	}

	return s
}
