package core

import (
	"log"
	"sync"

	"github.com/alikarimi999/shitcoin/core/types"
)

const (
	poolSize = 3
)

type pool interface {
	// this function can recieve transaction or mined block
	// if block mined by another node you must send block snapshot to prevent data race
	// if block mined by this node local must set true
	UpdatePool(o any, local bool)
	Handler()
}

type TxPool struct {
	Mu         *sync.Mutex
	c          *Chain
	wg         *sync.WaitGroup
	queueTxs   Transactions // verified transactions that recieved and didn't add to any block yet
	pendingTxs Transactions // verified transactions that are in mining block
	sealedTxs  Transactions // verified transactions that sealed in mining process

	queueCh    chan *types.Transaction
	sealCh     chan Transactions
	minedLocal chan bool
}

func NewTxPool(c *Chain) *TxPool {
	t := &TxPool{
		Mu:         &sync.Mutex{},
		c:          c,
		wg:         c.Wg,
		queueTxs:   make(Transactions),
		pendingTxs: make(Transactions),
		sealedTxs:  make(Transactions),
		queueCh:    make(chan *types.Transaction),
		sealCh:     make(chan Transactions),
		minedLocal: make(chan bool),
	}

	return t
}

func (tp *TxPool) Handler() {
	tp.wg.Add(1)
	defer tp.wg.Done()

	log.Println("Transaction Pool handler start!!!")
	for {
		select {
		case tx := <-tp.queueCh: // recieve from network
			tp.queueTxs[txid(tx.TxID)] = tx
			if tp.queIsFull() {
				go func() {
					tp.Mu.Lock()
					defer tp.Mu.Unlock()
					tp.pendingTxs = tp.queueTxs
					tp.c.Miner.Start(tp.queueTxs.convert())
					tp.queueTxs.clean()

					select {
					case local := <-tp.minedLocal:
						if local {
							tp.Mu.Lock()
							defer tp.Mu.Unlock()
							tp.pendingTxs.clean()
							return
						}

					default:
						go func() {
							tp.Mu.Lock()
							defer tp.Mu.Unlock()
							tp.sealedTxs = <-tp.sealCh
							tp.manageTxs()
						}()
					}
				}()
			}

		}
	}
}

// delete transactions that added to mined block before from queueTxs and pendingTxs
// and transfer pendingTxs that didn't added to mined block to queueTxs
func (tp *TxPool) manageTxs() {

	// delete used transactions
	for txid := range tp.sealedTxs {
		delete(tp.pendingTxs, txid)
		delete(tp.queueTxs, txid)
	}

	// merge pendingTxs and queueTxs
	for txid, tx := range tp.pendingTxs {
		tp.queueTxs[txid] = tx
	}

	tp.pendingTxs.clean()

}

// this function add transactions that added to mined block that recieved from other nodes
func (tp *TxPool) UpdatePool(o any, local bool) {

	switch t := o.(type) {
	case *types.Transaction:
		tp.queueCh <- t
	case *types.Block:
		if local {
			tp.minedLocal <- local
			return
		}
		tp.sealCh <- newTransations(t.Transactions)
		return
	default:
		return
	}

}

func (tp *TxPool) queIsFull() bool {
	return len(tp.queueTxs) == poolSize
}

func (t Transactions) convert() []*types.Transaction {
	txs := []*types.Transaction{}
	for _, tx := range t {
		txs = append(txs, tx)
	}
	return txs
}

func (t Transactions) clean() {
	t = make(Transactions)
}

func newTransations(txs []*types.Transaction) Transactions {
	t := make(Transactions)

	for _, tx := range txs {
		t[txid(tx.TxID)] = tx
	}

	return t
}
