package core

import (
	"errors"
	"fmt"
	"math"
	"math/big"
)

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

	return nil, errors.New(" ")
}
