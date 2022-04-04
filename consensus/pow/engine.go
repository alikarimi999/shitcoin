package pow

import (
	"bytes"
	"log"
	"math"
	"math/big"
	"sync/atomic"

	"github.com/alikarimi999/shitcoin/core/types"
)

const (
	Difficulty uint64 = 14
)

// PowEngine is a consensus engine based on proof-of-work alghorithm
type PowEngine struct {
	block  *types.Block
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

func NewPowEngine() *PowEngine {

	target := big.NewInt(1)
	target.Lsh(target, 256-uint(Difficulty))

	pe := &PowEngine{
		block:  types.NewBlock(),
		target: target,
		result: []byte{},
		pause:  make(chan struct{}),
		resume: make(chan struct{}),
		abort:  make(chan struct{}),
	}

	return pe
}

func (pe *PowEngine) mine() bool {

	var intHash big.Int

	var n uint64 = 0
	log.Printf("Start mining block %d\n", pe.block.BH.BlockIndex)
search:
	for n < math.MaxUint64 {

		select {
		case <-pe.abort:
			log.Printf("POWEngine nonce search for block %d aborted\n", pe.block.BH.BlockIndex)
			break search
		case <-pe.pause:
			log.Printf("POWEngine nonce search for block %d paused\n", pe.block.BH.BlockIndex)
			select {
			case <-pe.resume:
				log.Printf("POWEngine nonce search for block %d resumed\n", pe.block.BH.BlockIndex)
				continue search
			case <-pe.abort:
				log.Printf("POWEngine nonce search for block %d aborted\n", pe.block.BH.BlockIndex)
				break search
			}
		default:

			pe.block.BH.Nonce = n
			hash := pe.block.Hash()
			// fmt.Printf("\r%x", hash)
			intHash.SetBytes(hash)

			if intHash.Cmp(pe.target) == -1 {
				pe.block.BH.BlockHash = hash
				pe.result = hash
				atomic.StoreInt32(&pe.running, 0)
				return true

			}
			n++
		}

	}

	return false
}

func (pe *PowEngine) VerifyBlock(ch *types.ChainState, last_block types.Block) bool {

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

func (pe *PowEngine) Start(b *types.Block) bool {
	if !pe.IsRunning() {
		atomic.StoreInt32(&pe.running, 1)
		pe.block = b
		return pe.mine()
	}
	return false
}

func (pe *PowEngine) Pause() {
	if pe.IsRunning() {
		atomic.StoreInt32(&pe.running, 0)
		pe.pause <- struct{}{}
	}
}

func (pe *PowEngine) Close() {
	atomic.StoreInt32(&pe.running, 0)
	pe.abort <- struct{}{}
}

func (pe *PowEngine) Resume() {
	if !pe.IsRunning() {
		atomic.StoreInt32(&pe.running, 1)
		pe.resume <- struct{}{}
	}
}

func (pe *PowEngine) GetHash() []byte {
	return pe.result
}
