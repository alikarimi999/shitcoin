package core

import (
	"log"
	"math/big"
	"time"
)

type ProofOfWork struct {
	block  *Block
	target *big.Int
}

func NewProofOfWork(d uint64, b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, 256-uint(d))
	b.BH.Difficulty = d
	return &ProofOfWork{b, target}

}

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
