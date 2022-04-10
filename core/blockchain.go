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
	Mu *sync.Mutex
	Wg *sync.WaitGroup

	ChainId     types.Chainid
	ChainHeight uint64
	LastBlock   types.Block
	TxPool      pool
	ChainState  chainstate
	DB          database.Database
	Engine      consensus.Engin
	Validator   Validator
	MinerAdd    types.Address
	Miner       miner
	KnownNodes  map[types.NodeID]*types.Node
	DBPath      string
	Port        int
	MinedBlock  chan *types.Block
}

func NewChain(path string, port int, miner []byte) (*Chain, error) {
	c := &Chain{
		Mu: &sync.Mutex{},
		Wg: &sync.WaitGroup{},

		ChainId:     0,
		ChainHeight: 0,
		LastBlock:   *types.NewBlock(),

		Engine: pow.NewPowEngine(),

		MinerAdd:   miner,
		KnownNodes: make(map[types.NodeID]*types.Node),
		DBPath:     path,
		Port:       port,
		MinedBlock: make(chan *types.Block),
	}

	c.TxPool = NewTxPool(c)
	c.ChainState = NewState(filepath.Join(c.DBPath, "/chainstate"), c.Wg)
	c.Miner = NewMiner(c)
	c.Validator = NewValidator(c)

	c.DB.SetupDB(filepath.Join(c.DBPath, "/blocks"))

	return c, nil
}

func (c *Chain) SetupChain() error {
	c.creatGenesis()
	return nil
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
