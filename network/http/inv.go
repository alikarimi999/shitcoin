package network

import (
	"bytes"
	"fmt"
	"log"

	"github.com/labstack/echo/v4"
)

func (s *Server) SendInv(ctx echo.Context) error {

	gi := GetInv{}
	err := ctx.Bind(&gi)
	if err != nil {
		return err
	}

	inv := NewInv()
	inv.NodeId = s.Ch.Node.ID
	switch gi.InvType {
	case blockType:
		log.Printf("Node %s Requests for Block hashes\n", gi.NodeId)
		inv.InvType = blockType
		iter := s.Ch.NewIter()
		for {
			block, err := iter.Next()
			if err != nil {
				break
			}
			fmt.Printf("Adding block hash %x to inv\n", block.BH.BlockHash)
			inv.BlocksHash[blockIndex(block.BH.BlockIndex)] = block.BH.BlockHash
			inv.InvCount++

			if err != nil || bytes.Equal(block.BH.PrevHash, gi.LastHash) {
				break
			}
		}

	case txType:
		log.Printf("Node %s Requests for Transactions in transaction pool\n", gi.NodeId)
		inv.InvType = txType
		inv.TXs = append(inv.TXs, s.Ch.TxPool.GetPending()...)
		inv.TXs = append(inv.TXs, s.Ch.TxPool.GetQueue()...)
		inv.InvCount = len(inv.TXs)
	}

	ctx.JSONPretty(200, inv, " ")

	return nil

}
