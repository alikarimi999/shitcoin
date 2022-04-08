package core

import (
	"log"
	"sync"

	"github.com/alikarimi999/shitcoin/core/types"
)

type miner interface {
	Handler()
	Start(txs []*types.Transaction)
}

type Miner struct {
	wg      *sync.WaitGroup
	c       *Chain
	blockCh chan *types.Block
}

func NewMiner(c *Chain) *Miner {
	return &Miner{
		wg:      c.Wg,
		c:       c,
		blockCh: make(chan *types.Block),
	}
}

func (m *Miner) Handler() {
	m.wg.Add(1)
	defer m.wg.Done()
	log.Println("Miner Function start!")

	for {

		// Sender is in miner.Start Function
		b := <-m.blockCh
		m.c.Mu.Lock()
		b.BH.BlockIndex = m.c.ChainHeight
		b.BH.PrevHash = m.c.LastBlock.BH.BlockHash
		m.c.Mu.Unlock()

		b.BH.Miner = m.c.MinerAdd

		if m.c.Engine.Start(b) {
			log.Printf("Block %d with hash %x with %d transations Mined successfully\n", b.BH.BlockIndex, b.BH.BlockHash, len(b.Transactions))

			// reciver is in BroadMinedBlock function
			m.c.MinedBlock <- b.SnapShot()

			m.c.Mu.Lock()
			m.c.ChainHeight++
			m.c.LastBlock = *b
			m.c.Mu.Unlock()

			go m.c.State.StateTransition(b, true)
			go m.c.TxPool.UpdatePool(b, true)
			err := b.SaveBlockInDB(&m.c.DB, &sync.Mutex{})
			if err != nil {
				log.Printf("Block %x did not add to database\n\n", b.BH.BlockHash)
			}
			log.Printf("Block %x successfully added to database\n\n", b.BH.BlockHash)

		}
	}
}

func (m *Miner) Start(txs []*types.Transaction) {

	b := types.NewBlock()
	tx := MinerReward(m.c.MinerAdd, minerReward)
	m.c.State.StateTransition(tx.SnapShot(), false)
	m.c.TxPool.UpdatePool(tx.SnapShot(), false)
	b.Transactions = append(b.Transactions, tx)
	for _, tx := range txs {
		b.Transactions = append(b.Transactions, tx)
	}

	m.c.State.MineStarted(true)
	m.blockCh <- b
}
