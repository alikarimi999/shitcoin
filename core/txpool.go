package core

import (
	"log"
	"sort"
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
	ContinueHandler(cont bool)
	GetWaitGroup() *sync.WaitGroup
	GetQueue() []*types.Transaction
	GetPending() []*types.Transaction

	// GenesisUpdate(b *types.Block)
}

type TxPool struct {
	Mu *sync.Mutex
	c  *Chain
	wg *sync.WaitGroup

	WG         *sync.WaitGroup
	queueTxs   Transactions // verified transactions that recieved and didn't add to any block yet
	pendingTxs Transactions // verified transactions that are in mining block
	sealedTxs  Transactions // verified transactions that sealed in mining process

	queueCh       chan *types.Transaction
	minedRemoteCh chan Transactions
	continueCh    chan struct{}

	minedLocal chan bool
}

func NewTxPool(c *Chain) *TxPool {
	t := &TxPool{
		Mu:            &sync.Mutex{},
		c:             c,
		wg:            c.Wg,
		WG:            &sync.WaitGroup{},
		queueTxs:      make(Transactions),
		pendingTxs:    make(Transactions),
		sealedTxs:     make(Transactions),
		queueCh:       make(chan *types.Transaction),
		minedRemoteCh: make(chan Transactions),
		continueCh:    make(chan struct{}),
		minedLocal:    make(chan bool, 1),
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
			tp.Mu.Lock()
			tp.queueTxs[txid(tx.TxID)] = tx.SnapShot()
			tp.c.ChainState.StateTransition(tx, false)
			if tp.queIsFull() {
				tp.pendingTxs = tp.queueTxs
				tp.queueTxs = make(Transactions)
				// creat miner reward transaction
				mtx := MinerReward(tp.c.MinerAdd, minerReward)
				// tp.c.ChainState.StateTransition(mtx, false)
				tp.pendingTxs[txid(mtx.TxID)] = mtx

				txs := tp.pendingTxs.convert()

				// wait untile previous Mining proccess finish
				tp.WG.Wait()
				// notify chainstate handler that mining process is going to start
				tp.c.ChainState.MinerIsStarting(true)

				<-tp.continueCh // wait for ChainState Handler until take a snapshot from memSet

				// start mining process
				tp.c.Miner.Start(txs, tp.WG)

			}
			tp.Mu.Unlock()
		case local := <-tp.minedLocal:
			tp.Mu.Lock()

			if local {
				tp.pendingTxs = make(Transactions)

			}
			tp.Mu.Unlock()

		case tp.sealedTxs = <-tp.minedRemoteCh:
			tp.Mu.Lock()
			tp.manageTxs()

			// FIXME:
			// sending remaining transactions of transaction pool to state Handler
			txs := []*types.Transaction{}
			for _, tx := range tp.queueTxs {
				txs = append(txs, tx.SnapShot())
			}
			tp.c.ChainState.StateTransition(txs, false)
			tp.Mu.Unlock()

		}

	}
}

// delete transactions that added to remote mined block from queueTxs and pendingTxs
// and transfer pendingTxs that didn't added to mined block to queueTxs
func (tp *TxPool) manageTxs() {

	// delete used transactions
	for txid := range tp.sealedTxs {
		delete(tp.pendingTxs, txid)
		delete(tp.queueTxs, txid)
	}

	// merge pendingTxs and queueTxs
	for txid, tx := range tp.pendingTxs {
		if !tx.IsCoinbase() { // delete miner reward that remain in pendingTxs
			tp.queueTxs[txid] = tx
		}
	}

	tp.pendingTxs = make(Transactions)

}

func (tp *TxPool) ContinueHandler(cont bool) {
	if cont {
		tp.continueCh <- struct{}{}
	}
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
		tp.minedRemoteCh <- newTransations(t.Transactions)
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
	sort.SliceStable(txs, func(i, j int) bool { return txs[i].Timestamp < txs[j].Timestamp })
	return txs
}

func newTransations(txs []*types.Transaction) Transactions {
	t := make(Transactions)

	for _, tx := range txs {
		t[txid(tx.TxID)] = tx
	}

	return t
}

func (tp *TxPool) GetWaitGroup() *sync.WaitGroup {
	return tp.WG
}

func (tp *TxPool) GetQueue() []*types.Transaction {
	return tp.queueTxs.convert()
}

func (tp *TxPool) GetPending() []*types.Transaction {
	tp.Mu.Lock()
	defer tp.Mu.Unlock()
	return tp.pendingTxs.convert()
}

// func (tp *TxPool) GenesisUpdate(b *types.Block) {
// 	tp.
// }
