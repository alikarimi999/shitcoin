package core

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alikarimi999/shitcoin/consensus"
	"github.com/alikarimi999/shitcoin/core/types"
)

type miner interface {
	Handler()
	Start(txs []*types.Transaction, wg *sync.WaitGroup)
	IsRunning() bool
	MineGenesis(tx *types.Transaction)
	StartTime() time.Duration
}

type tmplBlock struct {
	block *types.Block
	wg    *sync.WaitGroup
}

type Miner struct {
	wg     *sync.WaitGroup
	c      *Chain
	engine consensus.Engin

	blockCh   chan *tmplBlock
	startTime time.Duration
}

func NewMiner(c *Chain) *Miner {
	return &Miner{
		wg:      c.Wg,
		c:       c,
		engine:  c.Engine,
		blockCh: make(chan *tmplBlock),
	}
}

func (m *Miner) Handler() {
	m.wg.Add(1)
	defer m.wg.Done()
	log.Println("Miner Function start!")

	for {

		select {
		case tmpl := <-m.blockCh: // Sender is in miner.Start Function
			go func(b *types.Block, wg *sync.WaitGroup) {
				defer wg.Done()
				m.c.Mu.Lock()
				b.BH.BlockIndex = atomic.LoadUint64(&m.c.ChainHeight)
				b.BH.PrevHash = m.c.LastBlock.BH.BlockHash
				b.BH.Miner = m.c.MinerAdd
				m.c.Mu.Unlock()
				time.Sleep(2 * time.Second)
				if m.engine.Start(b) {
					log.Printf("Block %d with hash %x with %d transations Mined successfully\n", b.BH.BlockIndex, b.BH.BlockHash, len(b.Transactions))

					// reciver is in BroadMinedBlock function
					m.c.MinedBlock <- b.SnapShot()

					atomic.AddUint64(&m.c.ChainHeight, 1)
					m.c.Mu.Lock()
					log.Printf("chain height is %d\n", atomic.LoadUint64(&m.c.ChainHeight))
					m.c.Node.NodeHeight++
					m.c.LastBlock = *b
					m.c.Node.LastHash = b.BH.BlockHash
					m.c.Mu.Unlock()

					m.c.ChainState.StateTransition(b, true)
					m.c.TxPool.UpdatePool(b, true)

					err := m.c.DB.SaveBlock(b, m.c.Node.ID, m.c.Node.ID, nil)
					if err != nil {
						log.Printf("Block %x did not add to database\n\n", b.BH.BlockHash)
						return
					}
					log.Printf("Block %x successfully added to database\n\n", b.BH.BlockHash)
					return
				}
			}(tmpl.block, tmpl.wg)

		}
	}
}

func (m *Miner) Start(txs []*types.Transaction, wg *sync.WaitGroup) {
	m.startTime = time.Duration(time.Now().UnixNano())
	b := types.NewBlock()

	for _, tx := range txs {
		b.Transactions = append(b.Transactions, tx)
	}

	wg.Add(1)
	m.blockCh <- &tmplBlock{
		block: b,
		wg:    wg,
	}

}

func (m *Miner) MineGenesis(tx *types.Transaction) {
	b := types.NewBlock()

	b.BH.BlockIndex = 0
	b.BH.PrevHash = m.c.LastBlock.BH.BlockHash
	b.BH.Miner = m.c.MinerAdd
	b.Transactions = append(b.Transactions, tx)

	if m.engine.Start(b) {
		m.c.ChainState.GenesisUpdate(b)
		m.c.Node.GenesisHash = b.BH.BlockHash

		m.c.ChainHeight++
		log.Printf("genesis block mined\nchain height is %d\n", atomic.LoadUint64(&m.c.ChainHeight))
		m.c.Node.NodeHeight++
		m.c.LastBlock = *b
		m.c.Node.LastHash = b.BH.BlockHash

		err := m.c.DB.SaveBlock(b, m.c.Node.ID, m.c.Node.ID, nil)
		if err != nil {
			log.Printf("Block %x did not add to database\n\n", b.BH.BlockHash)
			return
		}
		log.Printf("Block %x successfully added to database\n\n", b.BH.BlockHash)
		return

	}

}

func (m *Miner) IsRunning() bool {
	return m.engine.IsRunning()
}

func MinerReward(miner types.Address, amount int) *types.Transaction {
	pkh := types.Add2PKH(miner)
	reward := types.CoinbaseTx(pkh, amount)
	return reward
}

func (m *Miner) StartTime() time.Duration {
	return m.startTime
}
