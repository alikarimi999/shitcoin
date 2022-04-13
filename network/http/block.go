package network

import (
	"bytes"
	"fmt"
	"log"

	"github.com/alikarimi999/shitcoin/core"
	"github.com/alikarimi999/shitcoin/database"
	"github.com/labstack/echo/v4"
)

func (s *Server) MinedBlock(ctx echo.Context) error {

	mb := NewMsgBlock()
	err := ctx.Bind(mb)
	if err != nil {
		return err
	}

	for _, hash := range s.RecievedBlks {
		if bytes.Equal(hash, mb.Block.BH.BlockHash) {
			log.Printf("Block %x proccessed before\n", mb.Block.BH.BlockHash)
		}
	}

	log.Printf("Block: %d with hash %x mined by %s and received from Node %s\n", mb.Block.BH.BlockIndex, mb.Block.BH.BlockHash, mb.Miner, mb.Sender)
	s.RecievedBlks = append(s.RecievedBlks, mb.Block.BH.BlockHash)

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

		go s.Ch.ChainState.StateTransition(mb.Block, false)
		go s.Ch.TxPool.UpdatePool(mb.Block, false)

		s.Ch.LastBlock = *mb.Block
		s.Ch.ChainHeight++
		s.Ch.Node.NodeHeight++
		s.Ch.Node.LastHash = mb.Block.BH.BlockHash
		// Update NodeHeight of sender in KnownNodes
		s.Ch.KnownNodes[mb.Sender].NodeHeight++

		go s.Ch.AddBlockInDB(mb.Block, mb.Mu)

		// Broadcasting valid new Mined block in network
		// Reciver is BroadBlock function
		s.Client.BroadChan <- mb

	}

	return nil
}

// save block sender in chainstate database
func SaveBlocksenderInDB(hash []byte, sender string, d database.Database) error {

	err := d.DB.Put(hash, []byte(sender), nil)
	if err != nil {
		return err
	}
	log.Printf("Block: %x with sender %s saved in database\n", hash, sender)
	return nil
}
func (so *Server) SendBlock(ctx echo.Context) error {
	gb := new(GetBlock)
	err := ctx.Bind(gb)
	if err != nil {
		return err
	}
	hash := gb.BlockHash

	block := core.ReadBlock(so.Ch.DB, hash)
	mb := Msgblock(block, so.Ch.Node.ID, block.BH.Miner)

	fmt.Printf("\nNode %s wants Block %x\n", gb.Node, block.BH.BlockHash)
	ctx.JSONPretty(200, mb, " ")

	return nil
}
