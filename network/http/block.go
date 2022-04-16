package network

import (
	"bytes"
	"fmt"
	"log"
	"sort"
	"sync/atomic"

	"github.com/labstack/echo/v4"
)

func (s *Server) MinedBlock(ctx echo.Context) error {

	mb := NewMsgBlock()
	err := ctx.Bind(mb)
	if err != nil {
		return err
	}

	mb.Mu.Lock()
	for _, hash := range s.RecievedBlks {
		if bytes.Equal(hash, mb.Block.BH.BlockHash) {
			log.Printf("Block %x proccessed before\n", mb.Block.BH.BlockHash)
		}
	}

	log.Printf("Block: %d with hash %x mined by node %s and received from Node %s\n", mb.Block.BH.BlockIndex, mb.Block.BH.BlockHash, mb.Miner, mb.Sender)
	s.RecievedBlks = append(s.RecievedBlks, mb.Block.BH.BlockHash)
	mb.Mu.Unlock()

	s.Ch.Mu.Lock()
	defer s.Ch.Mu.Unlock()

	if mb.Block.BH.BlockIndex > s.Ch.LastBlock.BH.BlockIndex+2 {
		log.Println("Detecting fork")
		return nil
	}

	if mb.Block.BH.BlockIndex-1 == s.Ch.LastBlock.BH.BlockIndex {
		// pause mining process that trying to mine this block
		s.Ch.Engine.Pause()
		log.Println("  Proccessing Block")

		sort.SliceStable(mb.Block.Transactions, func(i, j int) bool { return mb.Block.Transactions[i].Timestamp < mb.Block.Transactions[j].Timestamp })

		if !s.Ch.Validator.ValidateBlk(mb.Block) {
			// resume paused mining process
			s.Ch.Engine.Resume()
			log.Printf("Block %x is not Valid\n", mb.Block.BH.BlockHash)
			return err

		}
		// stop mining process beacuse block mined by another node
		s.Ch.Engine.Close()
		fmt.Println()
		log.Printf("Block %x is valid\n", mb.Block.BH.BlockHash)

		s.Ch.ChainState.StateTransition(mb.Block, false)
		s.Ch.TxPool.UpdatePool(mb.Block, false)

		s.Ch.LastBlock = *mb.Block
		atomic.AddUint64(&s.Ch.ChainHeight, 1)
		s.Ch.Node.NodeHeight++
		s.Ch.Node.LastHash = mb.Block.BH.BlockHash
		// Update NodeHeight of sender in Peers
		s.Ch.Peers[mb.Sender].NodeHeight++

		s.Ch.DB.SaveBlock(mb.Block, mb.Sender, mb.Miner, nil)

		// Broadcasting valid new Mined block in network
		// Reciver is BroadBlock function
		s.Client.BroadChan <- mb

	}

	return nil
}

func (s *Server) SendBlock(ctx echo.Context) error {
	gb := new(GetBlock)
	err := ctx.Bind(gb)
	if err != nil {
		return err
	}
	hash := gb.BlockHash
	block, err := s.Ch.DB.GetBlockH(hash, nil)
	if err != nil {
		return err
	}
	mb := Msgblock(block, s.Ch.Node.ID, "")
	fmt.Printf("SendBlock: %x\n", mb.Block.BH.BlockHash)

	fmt.Printf("\nNode %s wants Block %x\n", gb.Node, block.BH.BlockHash)
	ctx.JSONPretty(200, mb, " ")

	return nil
}
