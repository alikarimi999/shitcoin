package core

import (
	"fmt"
	"log"
	"math/big"
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

func Mine(c *Chain, amount int) bool {

	b := NewBlock()
	b.BH.Miner = c.MinerAdd

	c.MemPool.TransferTxs2Block(b, amount)
	b.BH.BlockIndex = c.ChainHeight
	b.BH.PrevHash = c.LastBlock.BH.BlockHash
	pow := NewProofOfWork(10, b)

	_, err := pow.POW()
	if err != nil {
		fmt.Println("Finding nonce was unsuccesfl!!!")
		return false
	}
	fmt.Printf("Block %d mined with Hash: %x By %s\n", b.BH.BlockIndex, b.BH.BlockHash, b.BH.Miner)
	c.ChainHeight++
	c.LastBlock = b
	err = saveBlockInDB(*b, &c.DB)
	if err != nil {
		log.Fatalf("Block %x did not add to database\n\n", b.BH.BlockHash)
	}
	fmt.Printf("Block %x successfully added to database\n\n", b.BH.BlockHash)

	// Now we have to add utxoset to database
	c.SyncUtxoSet()

	return true

}
