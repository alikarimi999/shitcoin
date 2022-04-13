package network

import (
	"bytes"
	"fmt"
	"log"

	"github.com/alikarimi999/shitcoin/core/types"
	"github.com/labstack/echo/v4"
)

type MsgTX struct {
	SenderID string            `json:"sender"`
	TX       types.Transaction `json:"tx"`
}

func (s *Server) getTrx(ctx echo.Context) error {
	var mt MsgTX
	err := ctx.Bind(&mt)

	if err != nil {
		log.Panic(err)
	}

	for _, hash := range s.RecievedTxs {
		if bytes.Equal(hash, mt.TX.TxID) {
			log.Printf("Transaction %x proccessed before\n", mt.TX.TxID)
			return nil
		}
	}
	log.Printf("Transaction %x recieved from %s\n", mt.TX.TxID, mt.SenderID)

	s.RecievedTxs = append(s.RecievedTxs, mt.TX.TxID)

	if s.Ch.Validator.ValidateTX(&mt.TX) {
		s.Ch.TxPool.UpdatePool(&mt.TX, false)
		log.Printf("Transaction %x is valid\n", mt.TX.TxID)

		// Broadcast transaction
		BroadTrx(s.Ch, &mt)

		ctx.String(200, fmt.Sprintf("Transaction added to MemPool\n"))
		return nil

	}
	log.Printf("Transaction %x is not valid\n", mt.TX.TxID)
	ctx.String(200, "Transaction is not valid")
	return nil
}

func (s *Server) sendUTXOs(ctx echo.Context) error {

	account := ctx.QueryParam("account")
	msg := msgUTXOSet{
		Account: types.Account(account),
	}
	msg.Utxos = s.Ch.ChainState.GetTokens(msg.Account)

	ctx.JSONPretty(200, msg, "  ")
	return nil
}
