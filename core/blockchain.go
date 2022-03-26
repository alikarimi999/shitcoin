package core

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/alikarimi999/shitcoin/database"
)

const (
	BlockMaxTransactions = 4
)

type Chain struct {
	ChainId     Chainid
	ChainHeight uint64
	Blocks      []*Block
	LastBlock   *Block
	MemPool     *txsPool
	Chainstate  *ChainState
	DB          database.Database
	MinerAdd    Address
	KnownNodes  map[NodeID]*Node
	DBPath      string
	Port        int
}

type MsgBlock struct {
	Sender NodeID
	Block  *Block
	Miner  Address
}

func NewChain(path string, port int) (*Chain, error) {
	c := &Chain{
		ChainId:     0,
		ChainHeight: 0,
		Blocks:      make([]*Block, 0),
		LastBlock:   NewBlock(),

		MemPool: &txsPool{
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
	}
	c.DB.SetupDB(filepath.Join(c.DBPath, "/blocks"))
	c.MemPool.Chainstate.DB.SetupDB(filepath.Join(c.DBPath, "/chainstate"))
	return c, nil
}

func (c *Chain) SetupChain(miner Address, amount float64) error {

	err := c.creatGenesis(miner, amount)

	return err
}

func (c *Chain) Miner() {
	fmt.Println("Miner Function start!")
	cl := http.Client{Timeout: 5 * time.Second}
	for {
		if len(c.MemPool.Transactions) >= BlockMaxTransactions-1 {
			if Mine(c, 20) {
				// Broadcasting Mined block in network
				mb := NewMsgdBlock(c.LastBlock, NodeID(c.MinerAdd), c.MinerAdd)
				c.BroadBlock(mb, cl)

			}
		}
		time.Sleep(5 * time.Second)
	}

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
