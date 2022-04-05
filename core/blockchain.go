package core

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/alikarimi999/shitcoin/consensus"
	"github.com/alikarimi999/shitcoin/consensus/pow"
	"github.com/alikarimi999/shitcoin/core/types"
	"github.com/alikarimi999/shitcoin/database"
)

const (
	BlockMaxTransactions = 4
)

type Chain struct {
	Mu          *sync.Mutex
	ChainId     types.Chainid
	ChainHeight uint64
	LastBlock   types.Block
	MemPool     *types.MemPool
	Chainstate  *types.ChainState
	DB          database.Database
	Engine      consensus.Engin
	validator   Validator
	MinerAdd    types.Address
	KnownNodes  map[types.NodeID]*types.Node
	DBPath      string
	Port        int
	BlockChann  chan *types.Block
	MinedBlock  chan *types.Block
}

func NewChain(path string, port int) (*Chain, error) {
	c := &Chain{
		Mu:          &sync.Mutex{},
		ChainId:     0,
		ChainHeight: 0,
		LastBlock:   *types.NewBlock(),

		MemPool: &types.MemPool{
			Mu:           &sync.Mutex{},
			Transactions: []*types.Transaction{},
			Chainstate: &types.ChainState{
				Mu:    &sync.Mutex{},
				Utxos: make(map[types.Account][]*types.UTXO),
			},
		},
		Engine: pow.NewPowEngine(),
		Chainstate: &types.ChainState{
			Mu:    &sync.Mutex{},
			Utxos: make(map[types.Account][]*types.UTXO),
		},
		MinerAdd:   nil,
		KnownNodes: make(map[types.NodeID]*types.Node),
		DBPath:     path,
		Port:       port,
		BlockChann: make(chan *types.Block),
		MinedBlock: make(chan *types.Block),
	}
	c.DB.SetupDB(filepath.Join(c.DBPath, "/blocks"))
	c.Chainstate.DB.SetupDB(filepath.Join(c.DBPath, "/chainstate"))
	c.MemPool.Chainstate.DB = c.Chainstate.DB
	c.validator = NewBlockValidator(c, pow.NewPowEngine())

	return c, nil
}

func (c *Chain) SaveUtxoSet() error {

	saveUTXOsInDB(*c.MemPool.Chainstate)

	return nil
}

func (c *Chain) SetupChain(miner types.Address, amount float64) error {

	err := c.creatGenesis(miner, amount)

	return err
}

func (c *Chain) NewNode() *types.Node {

	n := &types.Node{
		NodeId:     types.NodeID(c.MinerAdd),
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
