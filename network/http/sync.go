package network

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/alikarimi999/shitcoin/core"
)

// Initial block download refers to the process where nodes synchronize themselves to the network
//by downloading blocks that are new to them
func IBD(o *Objects, cl http.Client) {

	// sync node is node with best chain
	syncNode := core.Node{}

	for _, node := range o.Ch.KnownNodes {
		if syncNode.NodeHeight <= node.NodeHeight {
			syncNode = node
		}
	}

	if syncNode.NodeHeight == o.Ch.ChainHeight {
		if !bytes.Equal(syncNode.LastHash, o.Ch.LastBlock.BH.BlockHash) {
			fmt.Printf(" Node \"%s\" and your node are not in same network please connect to another node\n", syncNode.FullAdd)
			delete(o.Ch.KnownNodes, syncNode.NodeId)
			return
		}
		fmt.Printf("Node is Synced with Node %s with Address %s\n", syncNode.NodeId, syncNode.FullAdd)
	}

	if syncNode.NodeHeight > o.Ch.ChainHeight {

		fmt.Printf("Sync node is %s with %d chain height\n", syncNode.FullAdd, syncNode.NodeHeight)
		fmt.Println("Trying to Sync with Sync Node")

		if o.Ch.ChainHeight == 0 {
			// Downloading genesis block
			block := getGen(syncNode.FullAdd, cl)
			// Updating UTXO Set base on genesis block transaction
			o.Ch.MemPool.Chainstate.UpdateUtxoSet(block.Transactions[0])
			// Save Genesis block in database
			err := core.SaveGenInDB(*block, &o.Ch.DB)
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Printf("Genesis Block added to database\n")
			o.Ch.LastBlock = block
			o.Ch.ChainHeight++
		}
		// Getting hash of remain mined Blocks from sync node
		bh, err := getData(o, syncNode.FullAdd, cl)
		if err != nil {
			fmt.Println(err.Error())
		}

		// Downloading mined Blocks
		for i := 0; i < len(bh); i++ {
			hash := bh[blockIndex(o.Ch.LastBlock.BH.BlockIndex+1)]

			block := getBlock(hash, syncNode.FullAdd, cl)
			if block == nil {
				break
			}
			fmt.Printf("Block %x Downloaded\n", block.BH.BlockHash)
			if !o.Ch.AddNewBlock(block) {
				break
			}
			o.Ch.LastBlock = block

		}
	}
}

func getBlock(hash []byte, syncAddress string, cl http.Client) *core.Block {
	data := GetBlock{hash}
	msg, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	resp, err := cl.Post(fmt.Sprintf("%s/getblock", syncAddress), "application/json", bytes.NewReader(msg))
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	b := new(core.Block)
	err = json.Unmarshal(body, b)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return b

}

// get genesis block
func getGen(syncNode string, cl http.Client) *core.Block {
	resp, err := cl.Get(fmt.Sprintf("%s/getgen", syncNode))
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	block := new(core.Block)
	json.Unmarshal(body, block)
	if block.Validate_hash() {
		fmt.Printf("Genesis Block downloaded\n")

		return block

	}
	fmt.Println("Genesis block is not valid")
	return nil

}

func getData(c *Objects, syncAddress string, cl http.Client) (map[blockIndex][]byte, error) {

	data := GetData{c.Ch.LastBlock.BH.BlockHash}
	b, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err.Error())
		return nil, nil
	}
	resp, err := cl.Post(fmt.Sprintf("%s/getdata", syncAddress), "application/json", bytes.NewReader(b))
	if err != nil {
		fmt.Println(err.Error())
		return nil, nil

	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		return nil, nil

	}
	inv := new(Inv)
	json.Unmarshal(body, inv)
	if inv.InvType == blockType {
		return inv.BlocksHash, nil
	}
	return nil, errors.New("sync node response currupt data to getData function")

}

// Get Node information from an address that running shitcoin client
func GetNodeInfo(fulladd string, cl http.Client) *NodeInfo {

	resp, err := cl.Get(fmt.Sprintf("%s/getnodeinfo", fulladd))
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	v := new(NodeInfo)
	json.Unmarshal(body, v)
	v.FullAdd = fulladd

	return v
}

// Add Node infromation that received from network to KnownNodes
func (n *NodeInfo) AddNode(c *core.Chain) {

	if _, ok := c.KnownNodes[n.NodeId]; !ok {

		c.KnownNodes[n.NodeId] = core.Node{
			NodeId:     n.NodeId,
			FullAdd:    n.FullAdd,
			LastHash:   n.LastHash,
			NodeHeight: n.NodeHeight,
		}
		fmt.Printf("Add Node %s with address %s to Known Nodes\n", n.NodeId, n.FullAdd)
		return
	}
	fmt.Printf("Node %s Exis in Known Nodes\n", n.NodeId)

}

// Creat a NodeInfo structure for broadcasting to network
func NewNodeInfo(c *core.Chain, port int) *NodeInfo {

	n := NodeInfo{
		Sender:     core.NodeID(c.MinerAdd),
		NodeId:     core.NodeID(c.MinerAdd),
		Port:       fmt.Sprintf("%d", port),
		FullAdd:    "",
		LastHash:   c.LastBlock.BH.BlockHash,
		NodeHeight: c.ChainHeight,
	}
	return &n
}

func (n *NodeInfo) BroadNode(c *core.Chain, cl http.Client) {

	for _, node := range c.KnownNodes {
		if node.NodeId == n.Sender || node.NodeId == n.NodeId {
			continue
		}

		// change NodeInfo sender to this Node NodeID
		n.Sender = core.NodeID(c.MinerAdd)
		b, _ := json.Marshal(n)

		fmt.Printf("Broadcast NodeInfo %s  with address (%s) to Node %s\n", n.NodeId, n.FullAdd, node.NodeId)
		cl.Post(fmt.Sprintf("%s/newnode", node.FullAdd), "application/json", bytes.NewReader(b))
	}

}
