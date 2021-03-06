package pow

import (
	"bytes"
	"log"
	"math"
	"math/big"
	"os"
	"strconv"
	"sync/atomic"

	"github.com/alikarimi999/shitcoin/core/types"
)

const (
	DefaultDifficulty uint64 = 14
)

// engine is a consensus engine based on proof-of-work alghorithm
type engine struct {
	block      *types.Block
	difficulty uint64
	target     *big.Int

	// block hash
	result []byte

	// atomic status counter
	running int32 // The indicator whether the consensus engine is running or not.

	// channels
	pause  chan struct{}
	resume chan struct{}
	abort  chan struct{}
}

func NewEngine() *engine {

	diff := get_diff()
	target := big.NewInt(1)
	target.Lsh(target, 256-uint(diff))

	pe := &engine{
		block:      types.NewBlock(),
		difficulty: diff,
		target:     target,
		result:     []byte{},
		pause:      make(chan struct{}),
		resume:     make(chan struct{}),
		abort:      make(chan struct{}),
	}

	return pe
}

func (e *engine) mine() bool {

	var intHash big.Int

	var n uint64 = 0
	log.Printf("Start mining block %d\n", e.block.BH.BlockIndex)
search:
	for n < math.MaxUint64 {

		select {
		case <-e.abort:
			log.Println("searching for nonce aborted")
			break search
		case <-e.pause:
			log.Println("searching for nonce paused")
			select {
			case <-e.resume:
				log.Println("searching for nonce resumed")
				continue search
			case <-e.abort:
				log.Println("searching for nonce aborted")
				break search
			}
		default:

			e.block.BH.Nonce = n
			hash := e.block.Hash()
			intHash.SetBytes(hash)

			if intHash.Cmp(e.target) == -1 {
				e.block.BH.BlockHash = hash
				e.result = hash
				atomic.StoreInt32(&e.running, 0)
				return true

			}
			n++
		}

	}

	return false
}

func (e *engine) VerifyBlock(b *types.Block, u *types.UtxoSet, last_block types.Block) bool {

	if b.BH.Difficulty == get_diff() && b.BH.BlockIndex-1 == last_block.BH.BlockIndex && bytes.Equal(b.BH.PrevHash, last_block.BH.BlockHash) && b.Validate_hash() {

		for _, tx := range b.Transactions {
			var account types.Account
			if tx.IsCoinbase() {
				account = types.Account(types.PK2Add(tx.TxOutputs[0].PublicKeyHash, true))
			} else {
				account = types.Account(types.PK2Add(tx.TxInputs[0].PublicKey, false))
			}
			if !tx.IsValid(u.Tokens[account]) {
				log.Printf("Transaction %x is invalid\n", tx.TxID)
				return false
			}
			u.UpdateUtxoSet(tx)
		}
		return true
	}
	return false
}

func (e *engine) IsRunning() bool {
	return atomic.LoadInt32(&e.running) == 1
}

func (e *engine) Start(b *types.Block) bool {
	if !e.IsRunning() {
		atomic.StoreInt32(&e.running, 1)
		e.block = b
		e.update_difficulty()
		return e.mine()
	}
	return false
}

func (e *engine) Pause() {
	if e.IsRunning() {
		atomic.StoreInt32(&e.running, 0)
		e.pause <- struct{}{}
	}
}

func (e *engine) Close() {
	atomic.StoreInt32(&e.running, 0)
	e.abort <- struct{}{}
}

func (e *engine) Resume() {
	if !e.IsRunning() {
		atomic.StoreInt32(&e.running, 1)
		e.resume <- struct{}{}
	}
}

func (e *engine) GetHash() []byte {
	return e.result
}

func (e *engine) update_difficulty() {
	diff := get_diff()
	target := big.NewInt(1)
	target.Lsh(target, 256-uint(diff))
	e.difficulty = diff
	e.block.BH.Difficulty = diff
}

func get_diff() uint64 {
	d := os.Getenv("DIFFICULTY")
	diff, err := strconv.ParseUint(d, 10, 64)
	if err != nil {
		return DefaultDifficulty
	}
	return diff
}
