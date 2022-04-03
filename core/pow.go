package core

import (
	"errors"
	"fmt"
	"math"
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

func (pow *ProofOfWork) POW() ([]byte, error) {

	var intHash big.Int

	var n uint64 = 0
	for n < math.MaxUint64 {

		pow.block.BH.Nonce = n
		hash := pow.block.Hash()
		fmt.Printf("\r%x", hash)
		intHash.SetBytes(hash[:])

		if intHash.Cmp(pow.target) == -1 {
			pow.block.BH.BlockHash = hash[:]
			fmt.Println()
			return hash[:], nil

		}
		n++

	}
	fmt.Println()

	return nil, errors.New(" ")
}
