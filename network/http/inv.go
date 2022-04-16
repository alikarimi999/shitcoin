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
	s.Ch.Mu.Lock()
	defer s.Ch.Mu.Unlock()
	inv := NewInv()
	inv.NodeId = s.Ch.Node.ID
	switch gi.InvType {
	case blockType:
		log.Printf("Node %s Requests for Block hashes\n", gi.NodeId)
		inv.InvType = blockType
		last_index := s.Ch.LastBlock.BH.BlockIndex
		inv.BlocksHash[blockIndex(last_index)] = s.Ch.LastBlock.BH.BlockHash
		inv.InvCount++

		for i := last_index - 1; ; i-- {
			hash, err := s.Ch.DB.GetBlkHash(i, nil)
			if err != nil || bytes.Equal(hash, gi.LastHash) {
				break
			}
			fmt.Printf("Adding block hash %x to inv\n", hash)
			inv.BlocksHash[blockIndex(i)] = hash
			inv.InvCount++
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
