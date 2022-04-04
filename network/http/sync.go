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
	"sync"
	"time"

	"github.com/alikarimi999/shitcoin/core"
	"github.com/alikarimi999/shitcoin/core/types"
)

const (
	MaxKnownNodes int = 8
	MaxNodesQueue
)

type NodesQueue struct {
	nodes chan *types.Node
}

func NewNodesQueue(size int) *NodesQueue {
	return &NodesQueue{
		nodes: make(chan *types.Node, size),
	}
}

func (n *NodesQueue) Push(node *types.Node) {
	select {
	case n.nodes <- node:
	default:
		// todo: right this shit
		fmt.Println("Nodes Queue is full")
	}
}

func (n *NodesQueue) Pop() *types.Node {
	select {
	case node := <-n.nodes:
		return node
	default:
		fmt.Println("Nodes Queue is empty ")
		return &types.Node{}
	}
}

// Initial block download refers to the process where nodes synchronize themselves to the network
//by downloading blocks that are new to them
func IBD(o *Objects, cl http.Client, wg sync.WaitGroup) {

	defer wg.Done()
	// sync node is node with best chain
	syncNode := &types.Node{}

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

		// Getting hash of remain mined Blocks from sync node
		bh, err := getData(o.Ch, syncNode.FullAdd, cl)
		if err != nil {
			fmt.Println(err.Error())
		}

		// Downloading mined Blocks
		for i := 0; i < len(bh); i++ {
			hash := bh[blockIndex(o.Ch.LastBlock.BH.BlockIndex+1)]

			mb := getBlock(hash, types.NodeID(o.Ch.MinerAdd), syncNode.FullAdd, cl)
			if reflect.DeepEqual(mb, new(MsgBlock)) {
				break
			}
			fmt.Printf("Block %x Downloaded from Node %s\n", mb.Block.BH.BlockHash, mb.Sender)

			// check if block is valid
			if BlockValidator(*mb.Block, o.Ch.MemPool.Chainstate, o.Ch.LastBlock) {
				fmt.Printf("Block %x is not valid\n", mb.Block.BH.BlockHash)
				break

			}
			fmt.Printf("Block %x is valid\n", mb.Block.BH.BlockHash)
			o.Ch.AddBlockInDB(mb.Block)
			o.Ch.SyncUtxoSet()

			o.Ch.LastBlock = *mb.Block

		}
	}
}

// sync with node that has a bigger chain
func Sync(c *core.Chain, n *types.Node) {
	if c.ChainHeight >= n.NodeHeight {
		fmt.Println(".... This Node does not have a better chain")
		return
	}

	cl := http.Client{Timeout: 5 * time.Second}
	// Getting hash of remain mined Blocks from sync node
	bh, err := getData(c, n.FullAdd, cl)
	if err != nil {
		fmt.Println(err.Error())
	}

	// Downloading mined Blocks
	for i := 0; i < len(bh); i++ {
		hash := bh[blockIndex(c.LastBlock.BH.BlockIndex+1)]

		mb := getBlock(hash, types.NodeID(c.MinerAdd), n.FullAdd, cl)
		if reflect.DeepEqual(mb, new(MsgBlock)) {
			break
		}
		fmt.Printf("... Block %x Downloaded from Node %s\n", mb.Block.BH.BlockHash, mb.Sender)

		// check if block is valid
		if BlockValidator(*mb.Block, c.MemPool.Chainstate, c.LastBlock) {
			fmt.Printf("... Block %x is not valid\n", mb.Block.BH.BlockHash)
			break

		}
		fmt.Printf("... Block %x is valid\n", mb.Block.BH.BlockHash)
		c.AddBlockInDB(mb.Block)
		c.SyncUtxoSet()

		c.LastBlock = *mb.Block

	}
	if c.ChainHeight == n.NodeHeight {
		fmt.Println(".....  Nodes Synced Now!  .....")
	}
}

func getBlock(hash []byte, nid types.NodeID, syncAddress string, cl http.Client) *MsgBlock {
	data := GetBlock{nid, hash}
	mb := new(MsgBlock)

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
func getGen(syncNode string, cl http.Client) *types.Block {
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
	block := new(types.Block)
	json.Unmarshal(body, block)
	if block.Validate_hash() {
		fmt.Printf("Genesis Block downloaded\n")

		return block

	}
	fmt.Println("Genesis block is not valid")
	return nil

}

func getData(c *core.Chain, syncAddress string, cl http.Client) (map[blockIndex][]byte, error) {

	data := GetData{c.LastBlock.BH.BlockHash}
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

func GetNewNodes(c *core.Chain, dst string, cl http.Client) []*types.Node {

	src_nodes := []*types.Node{}

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
			if n.NodeId == types.NodeID(c.MinerAdd) {
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

// Check if two nodes have same genesis block or not
// only if nodes have same genesis block can pair
// and return genesis block and a boolean
func IsInSameNet(genesis_hash []byte, node *types.Node) (*types.Block, bool) {
	gen_block := getGen(node.FullAdd, http.Client{Timeout: 5 * time.Second})
	return gen_block, bytes.Equal(gen_block.BH.BlockHash, genesis_hash)
}

func NodeInfo(dst string, cl http.Client) (*types.Node, error) {

	node := &types.Node{}
	resp, err := cl.Get(fmt.Sprintf("%s/nodeinfo", dst))
	if err != nil {
		return node, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return node, err
	}
	err = json.Unmarshal(body, node)
	if err != nil {
		return node, err
	}
	node.FullAdd = dst
	return node, nil

}

// this function pair two node in same network and download genesis block if it's needed
func PairNode(c *core.Chain, dst string) error {

	cl := http.Client{Timeout: 5 * time.Second}
	node, err := NodeInfo(dst, cl)
	if err != nil {
		log.Fatalln(err)
	}

	if c.ChainHeight >= 1 {
		hash, err := c.DB.DB.Get([]byte("genesis_block"), nil)
		if err != nil {
			log.Fatalln(err)
		}
		if _, ok := IsInSameNet(hash, node); !ok {
			log.Fatalf(`This two nodes don't have same Genesis Block\n
		If You want to connect to this network delete database "%s" and try again... `, c.DBPath)
		}
		// Nodes have same Genesis Block
		// so cane be paired
		c.KnownNodes[node.NodeId] = node
		fmt.Printf("...Node %s with address %s added to KnownNodes\n", node.NodeId, node.FullAdd)

		err = ShareNode(c, dst, cl)
		if err != nil {
			return err
		}
	}

	// Downloading genesis block
	block := getGen(node.FullAdd, cl)

	c.LastBlock = *block
	c.ChainHeight++

	// Updating UTXO Set base on genesis block transaction
	c.MemPool.Chainstate.UpdateUtxoSet(block.Transactions[0])
	// Save Genesis block in database
	err = core.SaveGenInDB(*block, &c.DB)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("Genesis Block added to database\n")

	c.SyncUtxoSet()

	err = ShareNode(c, dst, cl)
	if err != nil {
		return err
	}

	return nil
}

// Broadcast received transaction to Known Nodes
func BroadTrx(c *core.Chain, t types.Transaction) {
	cl := http.Client{Timeout: 5 * time.Second}
	b, err := json.Marshal(t)
	if err != nil {
		log.Println(err.Error())
		return
	}
	for _, n := range c.KnownNodes {
		log.Printf("Sending Transaction %x to Node %s\n", t.TxID, n.NodeId)
		cl.Post(fmt.Sprintf("%s/sendtrx", n.FullAdd), "application/json", bytes.NewReader(b))
	}
}

// This function Broadcast a new mined block in network
func (o *Objects) BroadMinedBlock() {

	for {
		// Sender is Miner function
		block := <-o.Ch.MinedBlock
		mb := MsgBlock{
			Sender: types.NodeID(o.Ch.MinerAdd),
			Block:  block,
			Miner:  o.Ch.MinerAdd,
		}

		for _, node := range o.Ch.KnownNodes {
			// dont send to miner of block or sender
			if mb.Sender == node.NodeId || types.NodeID(mb.Miner) == node.NodeId {
				continue
			}

			b, _ := json.Marshal(mb)
			log.Printf("Sending New Mined block %d: %x to %s\n", block.BH.BlockIndex, block.BH.BlockHash, node.NodeId)
			o.Cl.Post(fmt.Sprintf("%s/minedblock", node.FullAdd), "application/json", bytes.NewReader(b))
		}
	}
}

func (o *Objects) BroadBlock() {

	for {
		// Sender is MinedBlock function
		mb := <-o.BroadChan

		for _, node := range o.Ch.KnownNodes {
			// dont send to miner of block or sender
			if mb.Sender == node.NodeId || types.NodeID(mb.Miner) == node.NodeId {
				continue
			}

			prev_sender := mb.Sender

			// Set new sender
			mb.Sender = types.NodeID(o.Ch.MinerAdd)

			b, _ := json.Marshal(mb)
			fmt.Printf("Sending block %d: %x to %s which recieved from %s\n", mb.Block.BH.BlockIndex, mb.Block.BH.BlockHash, node.NodeId, prev_sender)
			o.Cl.Post(fmt.Sprintf("%s/minedblock", node.FullAdd), "application/json", bytes.NewReader(b))

		}

	}
}
