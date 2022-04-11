package network

import (
	"fmt"
	"log"

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
		o.Ch.TxPool.UpdatePool(t.SnapShot(), false)
		log.Printf("Transaction %x is valid\n", t.TxID)

		// Broadcast transaction
		BroadTrx(o.Ch, t)

		ctx.String(200, fmt.Sprintf("Transaction added to MemPool\n"))
		return nil

	}
	log.Printf("Transaction %x is not valid\n", t.TxID)
	ctx.String(200, "Transaction is not valid")
	return nil
}

func (o *Objects) sendUtxoset(ctx echo.Context) error {

	account := ctx.QueryParam("account")
	msg := msgUTXOSet{
		Account: types.Account(account),
	}
	msg.Utxos = o.Ch.ChainState.GetTokens(msg.Account)

	ctx.JSONPretty(200, msg, "  ")
	return nil
}
