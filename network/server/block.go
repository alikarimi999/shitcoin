package server

import (
	"bytes"
	"fmt"
	"log"
	"sort"
	"sync/atomic"

	netype "github.com/alikarimi999/shitcoin/network/types"
	"github.com/labstack/echo/v4"
)

func (s *Server) MinedBlock(ctx echo.Context) error {

	mb := netype.NewMsgBlock()
	err := ctx.Bind(mb)
	if err != nil {
		return err
	}

	for _, hash := range s.RecievedBlks {
		if bytes.Equal(hash, mb.Block.BH.BlockHash) {
			return nil
		}
	}

	log.Printf("block: %d with hash %x mined by node %s and received from Node %s\n", mb.Block.BH.BlockIndex, mb.Block.BH.BlockHash, mb.Miner, mb.Sender)
	s.RecievedBlks = append(s.RecievedBlks, mb.Block.BH.BlockHash)

	s.Ch.Mu.Lock()
	defer s.Ch.Mu.Unlock()

	if mb.Block.BH.BlockIndex > s.Ch.LastBlock.BH.BlockIndex+2 {
		log.Println("detecting fork")
		return nil
	}

	if mb.Block.BH.BlockIndex-1 == s.Ch.LastBlock.BH.BlockIndex {
		// pause mining process that trying to mine this block
		s.Ch.Engine.Pause()
		log.Println("  proccessing block")

		sort.SliceStable(mb.Block.Transactions, func(i, j int) bool { return mb.Block.Transactions[i].Timestamp < mb.Block.Transactions[j].Timestamp })

		if !s.Ch.Validator.ValidateBlk(mb.Block) {
			// resume paused mining process
			s.Ch.Engine.Resume()
			log.Printf("block %x is not Valid\n", mb.Block.BH.BlockHash)
			return err

		}
		// stop mining process beacuse block mined by another node
		s.Ch.Engine.Close()
		fmt.Println()
		log.Printf("block %x is valid\n", mb.Block.BH.BlockHash)

		s.Ch.ChainState.StateTransition(mb.Block, false)
		s.Ch.TxPool.UpdatePool(mb.Block, false)

		s.Ch.LastBlock = *mb.Block
		atomic.AddUint64(&s.Ch.ChainHeight, 1)
		s.Ch.Node.NodeHeight++
		s.Ch.Node.LastHash = mb.Block.BH.BlockHash
		// Update NodeHeight of sender in Peers
		s.PeerSet.Mu.Lock()
		s.PeerSet.Peers[mb.Sender].NodeHeight++
		s.PeerSet.Mu.Unlock()

		s.Ch.DB.SaveBlock(mb.Block, mb.Sender, mb.Miner, nil)

		s.BlkCh <- mb

	}

	return nil
}

func (s *Server) SendBlock(ctx echo.Context) error {
	gb := new(netype.GetBlock)
	err := ctx.Bind(gb)
	if err != nil {
		return err
	}
	hash := gb.BlockHash
	block, err := s.Ch.DB.GetBlockH(hash, nil)
	if err != nil {
		return err
	}
	mb := netype.Msgblock(block, s.Ch.Node.ID, "")
	log.Printf("node %s wants block %x\n", gb.Node, block.BH.BlockHash)
	ctx.JSONPretty(200, mb, " ")

	return nil
}

func (s *Server) GetMsgBlk() chan *netype.MsgBlock {
	return s.BlkCh
}
