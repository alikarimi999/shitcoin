package network

import (
	"sync"

	"github.com/alikarimi999/shitcoin/core"
	"github.com/alikarimi999/shitcoin/core/types"
	netype "github.com/alikarimi999/shitcoin/network/types"
)

type Client interface {
	IBD()
	Sync(n *types.Node)
	BroadTx(mt *netype.MsgTX)
	BroadBlock(mb *netype.MsgBlock)
}

type Server interface {
	Run(wg *sync.WaitGroup)
	GetMsgTx() chan *netype.MsgTX
	GetMsgBlk() chan *netype.MsgBlock
}

type Network struct {
	wg *sync.WaitGroup
	Client
	Server

	c *core.Chain
}

func Setup(cl Client, s Server, c *core.Chain, wg *sync.WaitGroup) {
	defer wg.Done()
	n := &Network{
		wg:     &sync.WaitGroup{},
		Client: cl,
		Server: s,
		c:      c,
	}

	n.wg.Add(3)
	go n.Server.Run(n.wg)
	go n.BroadBlock(n.wg)
	go n.BroadTx(n.wg)

	n.wg.Wait()
}

func (n *Network) BroadBlock(wg *sync.WaitGroup) {
	defer wg.Done()
	for {

		mb := netype.NewMsgBlock()
		select {
		case mb = <-n.Server.GetMsgBlk():

		case b := <-n.c.MinedBlockCh: // sender in miner function
			mb = &netype.MsgBlock{
				Sender: n.c.Node.ID,
				Miner:  n.c.Node.ID,
				Block:  b,
			}
		}
		n.Client.BroadBlock(mb)
	}
}

func (n *Network) BroadTx(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		mt := <-n.Server.GetMsgTx()
		n.Client.BroadTx(mt)
	}
}
