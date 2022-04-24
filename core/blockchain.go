package core

import (
	"path/filepath"
	"sync"

	"github.com/alikarimi999/shitcoin/config"
	"github.com/alikarimi999/shitcoin/consensus"
	"github.com/alikarimi999/shitcoin/consensus/pow"
	"github.com/alikarimi999/shitcoin/core/types"
	"github.com/alikarimi999/shitcoin/database"
)

const (
	BlockMaxTransactions = 4
)

// TODO: add memChain in Chain struct
type Chain struct {
	Mu *sync.Mutex

	ChainId     types.Chainid
	Node        *types.Node
	Config      config.Config
	ChainHeight uint64
	LastBlock   types.Block
	TxPool      pool
	ChainState  chainstate
	DB          database.DB
	Engine      consensus.Engin
	Validator   Validator
	MinerAdd    types.Address
	Miner       miner

	DBPath string
	Port   int

	MinedBlockCh chan *types.Block
}

func NewChain(path string, port int, miner []byte) (*Chain, error) {
	c := &Chain{
		Mu: &sync.Mutex{},

		ChainId:     0,
		ChainHeight: 0,
		LastBlock:   *types.NewBlock(),

		Engine: pow.NewEngine(),

		MinerAdd: miner,
		DBPath:   path,
		Port:     port,

		MinedBlockCh: make(chan *types.Block),
	}

	c.TxPool = NewTxPool(c)
	c.ChainState = NewState(c)
	c.Miner = NewMiner(c)
	c.Validator = NewValidator(c)
	c.Config = config.NewConfig(filepath.Join(c.DBPath, "/config.yaml"))
	c.DB = database.SetupDB(filepath.Join(c.DBPath, "/blocks"))
	c.Node = types.NewNode(c.Config, c.Port, c.LastBlock.BH.BlockHash, c.ChainHeight)

	return c, nil
}

func (c *Chain) SetupChain() error {
	c.creatGenesis()
	return nil
}
