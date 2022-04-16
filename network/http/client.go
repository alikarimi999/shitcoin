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
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alikarimi999/shitcoin/core"
	"github.com/alikarimi999/shitcoin/core/types"
)

type Client struct {
	Ch        *core.Chain
	Cl        http.Client
	BroadChan chan *MsgBlock
}

// Initial block download refers to the process where nodes synchronize themselves to the network
//by downloading blocks that are new to them
func (c *Client) IBD() {

	// sync node is node with best chain
	syncNode := &types.Node{}

	c.Ch.Mu.Lock()
	defer c.Ch.Mu.Unlock()

	for _, node := range c.Ch.Peers {
		if syncNode.NodeHeight <= node.NodeHeight {
			syncNode = node
		}
	}
	fmt.Printf("Sync Node IS %s\n", syncNode.ID)
	c.Sync(syncNode)
}

// sync with node that has a bigger chain
func (c *Client) Sync(n *types.Node) {

	// Getting hash of remain mined Blocks from sync node
	inv, err := getInv(blockType, c.Ch.Node.ID, c.Ch.LastBlock.BH.BlockHash, n.FullAdd, c.Cl)
	if err != nil {
		log.Fatal(err)
	}

	// Downloading mined Blocks
	// here we assumed map is sorted by blockIndex
	for i, hash := range inv.BlocksHash {
		if i != blockIndex(atomic.LoadUint64(&c.Ch.ChainHeight)) {
			break
		}
		fmt.Printf("Sync %x\n", hash)
		mb := getBlock(hash, c.Ch.Node.ID, n.FullAdd, c.Cl)
		if reflect.DeepEqual(mb, NewMsgBlock()) {
			break
		}
		fmt.Printf("... Block %x Downloaded from Node %s\n", mb.Block.BH.BlockHash, mb.Sender)

		// check if block is valid
		if c.Ch.Validator.ValidateBlk(mb.Block) {
			log.Printf("Block %d : %x is valid\n", mb.Block.BH.BlockIndex, mb.Block.BH.BlockHash)
			c.Ch.ChainState.StateTransition(mb.Block, false)
			c.Ch.TxPool.UpdatePool(mb.Block, false)

			c.Ch.DB.SaveBlock(mb.Block, mb.Sender, mb.Miner, nil)

			atomic.AddUint64(&c.Ch.ChainHeight, 1)
			c.Ch.LastBlock = *mb.Block
			c.Ch.Node.LastHash = mb.Block.BH.BlockHash

		}

	}

	//FIXME: dont use this mechanism
	height, err := getHeight(n.FullAdd)
	if err != nil {
		return
	}
	fmt.Printf("HEIGHT >>>> %d  %d\n", atomic.LoadUint64(&c.Ch.ChainHeight), height)
	if atomic.LoadUint64(&c.Ch.ChainHeight) == height {
		downloadTxPool(c.Ch, n.FullAdd)
		fmt.Println(".....  Nodes Synced Now!  .....")
	}
}

func getHeight(address string) (uint64, error) {
	resp, err := http.Get(fmt.Sprintf("%s/height", address))
	if err != nil {
		return 0, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	s, err := strconv.Atoi(string(body))
	if err != nil {
		return 0, err
	}
	return uint64(s), nil

}

// this function download transactions from sync node transaction pool
func downloadTxPool(c *core.Chain, dst string) {
	log.Println("Requesting transactions hashs of transaction pool")
	inv, err := getInv(txType, c.Node.ID, c.Node.LastHash, dst, http.Client{Timeout: 20 * time.Second})
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if inv.InvType != txType {
		log.Println("Incorrect data sended by sync node")
	}

	for _, tx := range inv.TXs {
		log.Printf("Transaction %x recieved from %s\n", tx.TxID, inv.NodeId)
		if c.Validator.ValidateTX(tx) {
			c.TxPool.UpdatePool(tx, false)
			log.Printf("Transaction %x is valid\n", tx.TxID)
			continue
		}
		log.Printf("Transaction %x is not valid\n", tx.TxID)
	}

}

func getBlock(hash []byte, nid string, syncAddress string, cl http.Client) *MsgBlock {
	data := GetBlock{nid, hash}
	mb := NewMsgBlock()

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
	fmt.Printf("getBlock %x\n", mb.Block.BH.BlockHash)

	// FIXME:
	mb.Mu = &sync.Mutex{}
	return mb

}

func getInv(invType InvType, nid string, lh []byte, syncAddress string, cl http.Client) (*Inv, error) {

	gi := GetInv{
		NodeId:   nid,
		InvType:  invType,
		LastHash: lh,
	}
	b, _ := json.Marshal(gi)

	resp, err := http.Post(fmt.Sprintf("%s/getinventory", syncAddress), "application/json", bytes.NewReader(b))
	if err != nil {
		log.Println(err.Error())
		return nil, errors.New("unsuccessful")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
		return nil, errors.New("unsuccessful")
	}
	inv := NewInv()
	json.Unmarshal(body, inv)
	return inv, nil

}

func GetNewNodes(c *core.Chain, dst string, cl http.Client) []*types.Node {

	src_nodes := []*types.Node{}

	// first element in slice always refer to node itself
	src_nodes = append(src_nodes, c.Node)

	for _, n := range c.Peers {
		src_nodes = append(src_nodes, n)
	}
	gn := &GetNode{src_nodes, nil}

	b, _ := json.Marshal(gn)
	resp, err := http.Post(fmt.Sprintf("%s/getnode", dst), "application/json", bytes.NewReader(b))
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

	// first DstNodes always is the node that share it's Peers with this node
	gn.ShareNodes[0].FullAdd = dst
	return gn.ShareNodes
}

// This function help to obtain some node addresses of network and add these Peers
func ShareNode(c *core.Chain, dst string, cl http.Client) error {

	// nodes that we didn't ask to share their nodes with us yet
	nq := NewNodesQueue(MaxNodesQueue)

Out:
	for i := 0; i <= MaxPeers; i++ {
		fmt.Printf("Requesting New Nodes from Node Address %s\n", dst)

		if len(c.Peers) >= MaxPeers {
			return fmt.Errorf("...this node has enough Peers")
		}
		share_nodes := GetNewNodes(c, dst, cl)
		if len(share_nodes) == 0 {
			fmt.Printf("...Node %s hadn't new node to share with us", dst)
			continue
		}

		for _, n := range share_nodes {

			fmt.Printf("...This Node %s Recieved from %s\n", n.ID, dst)

			if len(c.Peers) >= MaxPeers {
				break Out
			}

			// dont add if n refers to this node
			if n.ID == c.Node.ID {
				continue
			}
			if _, ok := c.Peers[n.ID]; ok {
				fmt.Println("...Node already exist")
				continue
			}

			c.Peers[n.ID] = n
			fmt.Printf("...Node %s with address %s added to Peers\n", n.ID, n.FullAdd)

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
func IsInSameNet(n1 *types.Node, n2 *types.Node) bool {
	return bytes.Equal(n1.GenesisHash, n2.GenesisHash)
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

	cl := http.Client{Timeout: 20 * time.Second}
	node, err := NodeInfo(dst, cl)
	if err != nil {
		log.Fatalln(err)
	}

	if atomic.LoadUint64(&c.ChainHeight) == 0 {
		// Downloading genesis block
		mb := getBlock(node.GenesisHash, node.ID, node.FullAdd, cl)

		c.LastBlock = *mb.Block
		atomic.AddUint64(&c.ChainHeight, 1)

		// update Node
		c.Node.GenesisHash = mb.Block.BH.BlockHash
		c.Node.LastHash = mb.Block.BH.BlockHash

		c.TxPool.UpdatePool(mb.Block, false)
		c.ChainState.StateTransition(mb.Block, false)

		// Save Genesis block in database
		err = c.DB.SaveBlock(mb.Block, mb.Sender, mb.Miner, nil)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("Genesis Block added to database\n")
	}

	err = ShareNode(c, dst, cl)
	if err != nil {
		return err
	}

	return nil
}

// Broadcast received transaction to  Peers
func BroadTrx(c *core.Chain, mt *MsgTX) {
	cl := http.Client{Timeout: 5 * time.Second}

	for _, n := range c.Peers {
		if mt.SenderID != n.ID {
			mt.SenderID = c.Node.ID
			b, _ := json.Marshal(mt)
			log.Printf("Sending Transaction %x to Node %s\n", mt.TX.TxID, n.ID)
			cl.Post(fmt.Sprintf("%s/sendtrx", n.FullAdd), "application/json", bytes.NewReader(b))
		}
	}
}

// This function Broadcast a new mined block in network
func (c *Client) BroadMinedBlock() {

	for {
		// Sender is Miner function
		block := <-c.Ch.MinedBlock
		mb := MsgBlock{
			Mu:     &sync.Mutex{},
			Sender: c.Ch.Node.ID,
			Block:  block,
			Miner:  c.Ch.Node.ID,
		}

		c.Ch.NMU.Lock()
		nodes := c.Ch.Peers

		for _, node := range nodes {
			// dont send to miner of block or sender
			if mb.Sender == node.ID || c.Ch.Node.ID == node.ID {
				continue
			}

			b, _ := json.Marshal(mb)
			log.Printf("Sending New Mined block %d: %x to %s\n", block.BH.BlockIndex, block.BH.BlockHash, node.ID)
			c.Cl.Post(fmt.Sprintf("%s/minedblock", node.FullAdd), "application/json", bytes.NewReader(b))
		}
		c.Ch.NMU.Unlock()
	}
}

func (c *Client) BroadBlock() {

	for {
		// Sender is MinedBlock function
		mb := <-c.BroadChan

		mb.Mu.Lock()
		defer mb.Mu.Unlock()
		for _, node := range c.Ch.Peers {
			// dont send to miner of block or sender
			if mb.Sender == node.ID || c.Ch.Node.ID == node.ID {
				continue
			}

			prev_sender := mb.Sender

			// Set new sender
			mb.Sender = c.Ch.Node.ID

			b, _ := json.Marshal(mb)
			fmt.Printf("Sending block %d: %x to %s which recieved from %s\n", mb.Block.BH.BlockIndex, mb.Block.BH.BlockHash, node.ID, prev_sender)
			c.Cl.Post(fmt.Sprintf("%s/minedblock", node.FullAdd), "application/json", bytes.NewReader(b))

		}

	}
}
