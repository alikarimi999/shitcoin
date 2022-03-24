package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/alikarimi999/shitcoin/database"
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
	KnownNodes  []string
	DBPath      string
}

type MinedBlock struct {
	Block *Block
	Miner Address
}

func NewChain(path string) (*Chain, error) {
	c := &Chain{
		Blocks:    make([]*Block, 0),
		LastBlock: &Block{},
		MemPool: &txsPool{
			Transactions: []*Transaction{},
			Chainstate: &ChainState{
				Utxos: make(map[Account][]*UTXO),
			},
		},
		Chainstate: &ChainState{
			Utxos: make(map[Account][]*UTXO),
		},
	}
	c.ChainId = 0
	c.DBPath = path
	c.DB.SetupDB(filepath.Join(c.DBPath, "/blocks"))
	c.MemPool.Chainstate.DB.SetupDB(filepath.Join(c.DBPath, "/chainstate"))
	return c, nil
}

func (c *Chain) SetupChain(miner Address, amount float64) error {

	err := c.creatGenesis(miner, amount)

	return err
}

func (c *Chain) Print() {
	fmt.Printf("\n%s Chain ID: %d  %s\n\n", strings.Repeat("=", 25), c.ChainId, strings.Repeat("=", 25))
	for _, b := range c.Blocks {
		b.Print()
	}

}

func (c *Chain) DeleteNode(node string) {

	s := []string{}

	for _, n := range c.KnownNodes {
		if n == node {
			continue
		}
		s = append(s, n)
	}
	c.KnownNodes = s

}

func (c *Chain) Miner() {
	fmt.Println("Miner Function start!")
	cl := http.Client{Timeout: 5 * time.Second}
	for {
		if len(c.MemPool.Transactions) >= 3 {
			if Mine(c, c.MinerAdd, 20) {
				mb := MinedBlock{c.LastBlock, c.MinerAdd}
				b, err := json.Marshal(mb)
				if err != nil {
					continue
				}
				for _, node := range c.KnownNodes {
					cl.Post(fmt.Sprintf("%s/minedblock", node), "application/json", bytes.NewReader(b))
				}
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func (c *Chain) NodeExist(n string) bool {
	for _, node := range c.KnownNodes {
		if node == n {
			return true
		}
	}
	return false
}
