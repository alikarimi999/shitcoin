package client

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
	"sync/atomic"

	"github.com/alikarimi999/shitcoin/core"
	"github.com/alikarimi999/shitcoin/core/types"
	netype "github.com/alikarimi999/shitcoin/network/types"
)

type Client struct {
	Ch *core.Chain
	Cl http.Client

	PeerSet *netype.PeerSet
}

// Initial block download refers to the process where nodes synchronize themselves to the network
//by downloading blocks that are new to them
func (c *Client) IBD() {

	// sync node is node with best chain
	syncNode := &types.Node{}

	c.PeerSet.Mu.Lock()
	defer c.PeerSet.Mu.Unlock()
	for _, node := range c.PeerSet.Peers {
		if syncNode.NodeHeight <= node.NodeHeight {
			syncNode = node
		}
	}
	fmt.Printf("Sync Node IS %s\n", syncNode.ID)
	c.Sync(syncNode)
}

// sync with node that has a bigger chain
func (c *Client) Sync(n *types.Node) {

	if atomic.LoadUint64(&c.Ch.ChainHeight) == 0 {
		c.getGen(n)
	}

	// Getting hash of remain mined Blocks from sync node
	inv, err := getInv(netype.BlockType, c.Ch.Node.ID, c.Ch.LastBlock.BH.BlockHash, n.FullAdd, c.Cl)
	if err != nil {
		log.Fatal(err)
	}

	// Downloading mined Blocks
	// here we assumed map is sorted by blockIndex
	for i, hash := range inv.BlocksHash {
		if i != netype.BlockIndex(atomic.LoadUint64(&c.Ch.ChainHeight)) {
			break
		}
		fmt.Printf("Sync %x\n", hash)
		mb := getBlock(hash, c.Ch.Node.ID, n.FullAdd, c.Cl)
		if reflect.DeepEqual(mb, netype.NewMsgBlock()) {
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

func getInv(invType netype.InvType, nid string, lh []byte, syncAddress string, cl http.Client) (*netype.Inv, error) {

	gi := netype.GetInv{
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
	inv := netype.NewInv()
	json.Unmarshal(body, inv)
	return inv, nil

}

// Check if two nodes have same genesis block or not
// only if nodes have same genesis block can pair
// and return genesis block and a boolean
func IsInSameNet(n1 *types.Node, n2 *types.Node) bool {
	return bytes.Equal(n1.GenesisHash, n2.GenesisHash)
}
