package core

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/alikarimi999/shitcoin/database"
)

const (
	BlockMaxTransactions = 4
)

type Chain struct {
	Mu          sync.Mutex
	ChainId     Chainid
	ChainHeight uint64
	Blocks      []*Block
	LastBlock   Block
	MemPool     *memPool
	Chainstate  *ChainState
	DB          database.Database
	MinerAdd    Address
	KnownNodes  map[NodeID]*Node
	DBPath      string
	Port        int
	BlockChann  chan *Block
	MinedBlock  chan *Block
}

func NewChain(path string, port int) (*Chain, error) {
	c := &Chain{
		Mu:          sync.Mutex{},
		ChainId:     0,
		ChainHeight: 0,
		Blocks:      make([]*Block, 0),
		LastBlock:   *NewBlock(),

		MemPool: &memPool{
			Transactions: []*Transaction{},
			Chainstate: &ChainState{
				Utxos: make(map[Account][]*UTXO),
			},
		},
		Chainstate: &ChainState{
			Utxos: make(map[Account][]*UTXO),
		},
		MinerAdd:   nil,
		KnownNodes: make(map[NodeID]*Node),
		DBPath:     path,
		Port:       port,
		BlockChann: make(chan *Block),
		MinedBlock: make(chan *Block),
	}
	c.DB.SetupDB(filepath.Join(c.DBPath, "/blocks"))
	c.Chainstate.DB.SetupDB(filepath.Join(c.DBPath, "/chainstate"))
	c.MemPool.Chainstate.DB = c.Chainstate.DB
	return c, nil
}

func (c *Chain) SetupChain(miner Address, amount float64) error {

	err := c.creatGenesis(miner, amount)

	return err
}

func (c *Chain) NewNode() *Node {

	n := &Node{
		NodeId:     NodeID(c.MinerAdd),
		FullAdd:    "",
		Port:       fmt.Sprintf(":%d", c.Port),
		LastHash:   c.LastBlock.BH.BlockHash,
		NodeHeight: c.ChainHeight,
	}

	return n
}

func (c *Chain) SnapShot() *Chain {
	ch := c
	ch.Chainstate = c.Chainstate.SnapShot()
	ch.MemPool = c.MemPool.SnapShot()
	return ch
}
