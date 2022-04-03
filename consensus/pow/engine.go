package pow

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"math"
	"math/big"
	"sync/atomic"

	"github.com/alikarimi999/shitcoin/core"
)

// PowEngine is a consensus engine based on proof-of-work alghorithm
type PowEngine struct {
	block  *core.Block
	target *big.Int

	// block hash
	result []byte

	// atomic status counter
	running int32 // The indicator whether the consensus engine is running or not.

	// channels
	pause  chan struct{}
	resume chan struct{}
	abort  chan struct{}
}

func NewPowEngine(d uint64, b *core.Block) *PowEngine {

	target := big.NewInt(1)
	target.Lsh(target, 256-uint(d))
	b.BH.Difficulty = d

	pe := &PowEngine{
		block:  b,
		target: target,
		result: b.BH.BlockHash,
		pause:  make(chan struct{}),
		resume: make(chan struct{}),
		abort:  make(chan struct{}),
	}

	return pe
}

func (pe *PowEngine) mine() ([]byte, error) {

	var intHash big.Int

	var n uint64 = 0

search:
	for n < math.MaxUint64 {

		select {
		case <-pe.abort:
			log.Println("POWEngine nonce search aborted")
			break search
		case <-pe.pause:
			log.Println("POWEngine nonce search paused")
			select {
			case <-pe.resume:
				log.Println("POWEngine nonce search resumed")
				continue search
			case <-pe.abort:
				log.Println("POWEngine nonce search aborted")
				break search
			}
		default:

			pe.block.BH.Nonce = n
			hash := pe.block.Hash()
			fmt.Printf("\r%x", hash)
			intHash.SetBytes(hash)

			if intHash.Cmp(pe.target) == -1 {
				pe.block.BH.BlockHash = hash
				pe.result = hash
				return hash, nil

			}
			n++
		}

	}
	fmt.Println()

	return nil, errors.New(" ")
}

func (pe *PowEngine) VerifyBlock(ch *core.ChainState, last_block core.Block) bool {

	b := *pe.block

	if b.BH.BlockIndex-1 == last_block.BH.BlockIndex && bytes.Equal(b.BH.PrevHash, last_block.BH.BlockHash) && b.Validate_hash() {

		if utxos, valid := ch.Validate_blk_trx(b); valid {
			ch.Utxos = utxos
			return true
		}

		return false
	}
	return false
}

func (pe *PowEngine) IsRunning() bool {
	return atomic.LoadInt32(&pe.running) == 1
}

func (pe *PowEngine) Start() ([]byte, error) {
	atomic.StoreInt32(&pe.running, 1)
	hash, err := pe.mine()

	return hash, err
}

func (pe *PowEngine) Pause() {
	if !pe.IsRunning() {
		atomic.StoreInt32(&pe.running, 0)
		pe.pause <- struct{}{}
	}
}

func (pe *PowEngine) Abort() {
	atomic.StoreInt32(&pe.running, 0)
	close(pe.abort)
}

func (pe *PowEngine) Resume() {
	if pe.IsRunning() {
		atomic.StoreInt32(&pe.running, 1)
		pe.resume <- struct{}{}
	}
}
