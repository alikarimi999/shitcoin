package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/alikarimi999/shitcoin/core/types"
	netype "github.com/alikarimi999/shitcoin/network/types"
)

func (c *Client) BroadBlock(mb *netype.MsgBlock) {

	c.PeerSet.Mu.Lock()
	defer c.PeerSet.Mu.Unlock()

	for _, node := range c.PeerSet.Peers {
		// dont send to miner of block or sender
		if mb.Sender == node.ID || c.Ch.Node.ID == node.ID {
			continue
		}

		prev_sender := mb.Sender

		// Set new sender
		mb.Sender = c.Ch.Node.ID

		b, _ := json.Marshal(mb)
		fmt.Printf("sending block %d: %x to %s which recieved from %s\n", mb.Block.BH.BlockIndex, mb.Block.BH.BlockHash, node.ID, prev_sender)
		c.Cl.Post(fmt.Sprintf("%s/minedblock", node.FullAdd), "application/json", bytes.NewReader(b))

	}
}

func getBlock(hash []byte, nid string, syncAddress string, cl http.Client) *netype.MsgBlock {
	data := netype.GetBlock{
		Node:      nid,
		BlockHash: hash,
	}
	mb := netype.NewMsgBlock()

	msg, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err.Error())
		return mb
	}
	resp, err := cl.Post(fmt.Sprintf("%s/getblock", syncAddress), "application/json", bytes.NewReader(msg))
	if err != nil {
		fmt.Println(err.Error())
		return mb
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		return mb
	}
	err = json.Unmarshal(body, mb)
	if err != nil {
		fmt.Println(err.Error())
		return mb
	}

	return mb

}

// Downloading genesis block
func (c *Client) getGen(n *types.Node) {
	if atomic.LoadUint64(&c.Ch.ChainHeight) == 0 {
		mb := getBlock(n.GenesisHash, n.ID, n.FullAdd, c.Cl)

		c.Ch.LastBlock = *mb.Block
		atomic.AddUint64(&c.Ch.ChainHeight, 1)

		// update Node
		c.Ch.Node.GenesisHash = mb.Block.BH.BlockHash
		c.Ch.Node.LastHash = mb.Block.BH.BlockHash

		c.Ch.TxPool.UpdatePool(mb.Block, false)
		c.Ch.ChainState.StateTransition(mb.Block, false)

		// Save Genesis block in database
		err := c.Ch.DB.SaveBlock(mb.Block, mb.Sender, mb.Miner, nil)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("genesis block added to database\n")
	}
}
