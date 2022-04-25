package server

import (
	"bytes"
	"fmt"
	"log"

	"github.com/alikarimi999/shitcoin/core/types"
	netype "github.com/alikarimi999/shitcoin/network/types"
	"github.com/labstack/echo/v4"
)

func (s *Server) getTrx(ctx echo.Context) error {
	var mt netype.MsgTX
	err := ctx.Bind(&mt)

	if err != nil {
		log.Panic(err)
	}

	for _, hash := range s.RecievedTxs {
		if bytes.Equal(hash, mt.TX.TxID) {
			return nil
		}
	}
	log.Printf("Transaction %x recieved from %s\n", mt.TX.TxID, mt.SenderID)

	s.RecievedTxs = append(s.RecievedTxs, mt.TX.TxID)
	if s.Ch.Validator.ValidateTX(&mt.TX) {
		s.Ch.TxPool.UpdatePool(&mt.TX, false)
		log.Printf("Transaction %x is valid\n", mt.TX.TxID)

		// reciever in Protocol BroadTrx
		s.TxCh <- &mt

		ctx.String(200, fmt.Sprintf("Transaction added to MemPool\n"))
		return nil

	}
	log.Printf("Transaction %x is not valid\n", mt.TX.TxID)
	ctx.String(200, "Transaction is not valid")
	return nil
}

func (s *Server) sendUTXOs(ctx echo.Context) error {

	account := ctx.QueryParam("account")
	msg := netype.MsgUTXOSet{
		Account: types.Account(account),
	}
	msg.Utxos = s.Ch.ChainState.GetTokens(msg.Account)

	ctx.JSONPretty(200, msg, "  ")
	return nil
}

func (s *Server) GetMsgTx() chan *netype.MsgTX {
	return s.TxCh
}
