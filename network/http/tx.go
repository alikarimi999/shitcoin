package network

import (
	"log"

	"github.com/alikarimi999/shitcoin/core"
	"github.com/labstack/echo/v4"
)

func (h *Objects) getTrx(ctx echo.Context) error {
	c := h.Ch
	var t core.Transaction
	err := ctx.Bind(&t)

	if err != nil {
		log.Panic(err)
	}

	err = c.AddTx2Pool(&t)
	if err != nil {
		return ctx.String(200, err.Error())
	}
	return ctx.String(200, "Transaction added to MemPool")
}

func (h *Objects) sendUtxoset(ctx echo.Context) error {

	account := ctx.QueryParam("account")
	msg := sendUtxoset(h.Ch, core.Account(account))

	ctx.JSONPretty(200, msg, "  ")
	return nil
}

func sendUtxoset(c *core.Chain, a core.Account) msgUTXOSet {

	var s msgUTXOSet
	s.Account = a
	utxos := c.MemPool.Chainstate.Utxos[a]

	for _, utxo := range utxos {
		// su := core.UTXO{utxo.Txid, utxo.Index, utxo.Txout}
		s.Utxos = append(s.Utxos, *utxo)
	}

	return s
}
