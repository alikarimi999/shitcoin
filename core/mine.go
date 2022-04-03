package core

import (
	"log"
	"time"
)

func Mine(c *Chain, b *Block, amount int) bool {

	log.Println("Start Mining")

	pow := NewProofOfWork(15, b)
	pow.block.BH.Timestamp = time.Now().UnixNano()
	_, err := pow.POW()
	if err != nil {
		log.Println("Finding nonce was unsuccesfl!!!")
		return false
	}
	log.Printf("Block %d mined with Hash: %x By %s At %d moment\n", b.BH.BlockIndex, b.BH.BlockHash, b.BH.Miner, b.BH.Timestamp)

	return true

}
