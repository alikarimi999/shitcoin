package network

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"

	"github.com/alikarimi999/shitcoin/core"
)

const (
	MaxKnownNodes int = 8
	MaxNodesQueue
)

type NodesQueue struct {
	nodes chan *core.Node
}

func NewNodesQueue(size int) *NodesQueue {
	return &NodesQueue{
		nodes: make(chan *core.Node, size),
	}
}

func (n *NodesQueue) Push(node *core.Node) {
	select {
	case n.nodes <- node:
	default:
		// todo: right this shit
		fmt.Println("Nodes Queue is full")
	}
}

func (n *NodesQueue) Pop() *core.Node {
	select {
	case node := <-n.nodes:
		return node
	default:
		fmt.Println("There is not any other node for requesting new node ")
		return &core.Node{}
	}
}

// Initial block download refers to the process where nodes synchronize themselves to the network
//by downloading blocks that are new to them
func IBD(o *Objects, cl http.Client) {

	// sync node is node with best chain
	syncNode := &core.Node{}

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

			o.Ch.LastBlock = block
			o.Ch.ChainHeight++

			// Updating UTXO Set base on genesis block transaction
			o.Ch.MemPool.Chainstate.UpdateUtxoSet(block.Transactions[0])
			// Save Genesis block in database
			err := core.SaveGenInDB(*block, &o.Ch.DB)
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Printf("Genesis Block added to database\n")

		}
		// Getting hash of remain mined Blocks from sync node
		bh, err := getData(o, syncNode.FullAdd, cl)
		if err != nil {
			fmt.Println(err.Error())
		}

		// Downloading mined Blocks
		for i := 0; i < len(bh); i++ {
			hash := bh[blockIndex(o.Ch.LastBlock.BH.BlockIndex+1)]

			mb := getBlock(hash, core.NodeID(o.Ch.MinerAdd), syncNode.FullAdd, cl)
			if reflect.DeepEqual(mb, new(core.MsgBlock)) {
				break
			}
			fmt.Printf("Block %x Downloaded from Node %s\n", mb.Block.BH.BlockHash, mb.Sender)

			// check if block is valid
			if !o.Ch.BlockValidator(*mb.Block) {
				fmt.Printf("Block %x is not valid\n", mb.Block.BH.BlockHash)
				break

			}
			fmt.Printf("Block %x is valid\n", mb.Block.BH.BlockHash)
			o.Ch.AddBlockInDB(mb.Block)
			o.Ch.SyncUtxoSet()

			o.Ch.LastBlock = mb.Block

		}
	}
}

func getBlock(hash []byte, nid core.NodeID, syncAddress string, cl http.Client) *core.MsgBlock {
	data := GetBlock{nid, hash}
	mb := new(core.MsgBlock)

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

func GetNewNodes(c *core.Chain, dst string, cl http.Client) []*core.Node {

	src_nodes := []*core.Node{}

	// first element in slice always refer to node itself
	src_nodes = append(src_nodes, c.NewNode())

	for _, n := range c.KnownNodes {
		src_nodes = append(src_nodes, n)
	}
	gn := &GetNode{src_nodes, nil}

	b, _ := json.Marshal(gn)
	resp, err := cl.Post(fmt.Sprintf("%s/getnode", dst), "application/json", bytes.NewReader(b))
	if err != nil {
		fmt.Println(err.Error())
		return src_nodes
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		return src_nodes
	}
	gn = new(GetNode)
	err = json.Unmarshal(body, gn)
	if err != nil {
		fmt.Println(err.Error())
		return src_nodes
	}

	// first DstNodes always is the node that share it's known nodes with this node
	gn.ShareNodes[0].FullAdd = dst
	return gn.ShareNodes
}

// This function help to obtain some node addresses of network and add these KnownNodes
func ShareNode(c *core.Chain, dst string, cl http.Client) error {

	// nodes that we didn't ask to share their nodes with us yet
	nq := NewNodesQueue(MaxNodesQueue)

Out:
	for i := 0; i <= MaxKnownNodes; i++ {
		fmt.Printf("Requesting New Nodes from Node Address %s\n", dst)

		if len(c.KnownNodes) >= MaxKnownNodes {
			return fmt.Errorf("...this node has enough known node")
		}
		share_nodes := GetNewNodes(c, dst, cl)
		if len(share_nodes) == 0 {
			fmt.Printf("...Node %s hadn't new node to share with us", dst)
			continue
		}

		for _, n := range share_nodes {

			fmt.Printf("...This Node %s Recieved from %s\n", n.NodeId, dst)

			if len(c.KnownNodes) >= MaxKnownNodes {
				break Out
			}

			// dont add if n refers to this node
			if n.NodeId == core.NodeID(c.MinerAdd) {
				continue
			}
			if _, ok := c.KnownNodes[n.NodeId]; ok {
				fmt.Println("...Node already exist")
				continue
			}

			c.KnownNodes[n.NodeId] = n
			fmt.Printf("...Node %s with address %s added to KnownNodes\n", n.NodeId, n.FullAdd)

			// dont send getnode to previous destination node again
			if n.FullAdd != dst {
				nq.Push(n)
			}
		}

		// get new nodes from nodes that sends by previous nodes

		dst = nq.Pop().FullAdd
		if dst == "" {
			break
		}

	}

	return nil
}
