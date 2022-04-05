package network

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/alikarimi999/shitcoin/core"
	"github.com/alikarimi999/shitcoin/core/types"
	"github.com/alikarimi999/shitcoin/database"
	"github.com/labstack/echo/v4"
)

const (
	NETWORKPROTO string = "http://"
)

type Objects struct {
	Mu        sync.Mutex
	Ch        *core.Chain
	Port      int
	BroadChan chan *MsgBlock
	Cl        http.Client
}

func RunServer(o *Objects, port int) {

	go o.BroadMinedBlock()
	go o.BroadBlock()

	e := echo.New()

	e.GET("/getutxo", o.sendUtxoset)
	e.POST("sendtrx", o.getTrx)
	e.GET("/getgen", o.SendGen)
	e.POST("getdata", o.SendInv)
	e.POST("/getblock", o.SendBlock)
	e.POST("/minedblock", o.MinedBlock)
	e.POST("/getnode", o.SendNodes)
	e.GET("nodeinfo", o.SendNodeInfo)
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", port)))
}

func (o *Objects) SendNodeInfo(ctx echo.Context) error {
	node := o.Ch.NewNode()
	err := ctx.Bind(node)
	if err != nil {
		return err
	}
	ctx.JSONPretty(200, node, " ")
	return nil
}

func (o *Objects) SendNodes(ctx echo.Context) error {
	gn := &GetNode{}

	err := ctx.Bind(gn)
	if err != nil {
		return err
	}

	sender := NETWORKPROTO + ctx.RealIP() + gn.SrcNodes[0].Port
	senderID := gn.SrcNodes[0].NodeId
	gn.SrcNodes[0].FullAdd = sender
	fmt.Printf("Node %s with Address %s Requesitng new node\n", gn.SrcNodes[0].NodeId, gn.SrcNodes[0].FullAdd)

	gn.ShareNodes = sendNode(o.Ch, gn.SrcNodes, senderID)

	for _, n := range gn.SrcNodes {
		if len(o.Ch.KnownNodes) >= MaxKnownNodes {
			break
		}
		if _, ok := o.Ch.KnownNodes[n.NodeId]; !ok && n.NodeId != types.NodeID(o.Ch.MinerAdd) {
			o.Ch.KnownNodes[n.NodeId] = n
			fmt.Printf("...Add Node %s with address %s to KnownNodes\n", n.NodeId, n.FullAdd)
		}

		if o.Ch.ChainHeight < n.NodeHeight {
			fmt.Printf("... Node %s had %d mined block more\n", n.NodeId, n.NodeHeight-o.Ch.ChainHeight)
			fmt.Printf("... Trying to sync with Node %s\n", n.NodeId)
			o.Ch.Mu.Lock()
			Sync(o.Ch, n)
			o.Ch.Mu.Unlock()
		}
	}

	ctx.JSONPretty(200, gn, " ")
	return nil
}

func sendNode(c *core.Chain, src []*types.Node, sender types.NodeID) []*types.Node {
	share_nodes := []*types.Node{}

	// first node that any node share to other nodes refers to itself
	n := c.NewNode()
	share_nodes = append(share_nodes, n)
	fmt.Printf("...Sending Node %s\n", n.NodeId)

Out:
	for _, node := range c.KnownNodes {

		// every node share 1 KnownNodes an itself address
		// this made applicant node to requests other nodes for sharing their nodes too
		// and this made the network more destributed
		if len(share_nodes) >= 2 {
			break
		}

		// dont share node if applicant node already has it
		for _, n := range src {
			if node.NodeId == n.NodeId {
				fmt.Printf("... Node %s already has this Node %s with address %s so don't send it\n", sender, n.NodeId, n.FullAdd)
				continue Out
			}
		}
		if node.NodeId == sender {
			continue Out
		}
		share_nodes = append(share_nodes, node)
		fmt.Printf("...Sending Node %s with Address %s\n", node.NodeId, node.FullAdd)

	}
	return share_nodes
}

func (o *Objects) MinedBlock(ctx echo.Context) error {

	mb := NewMsgBlock()
	err := ctx.Bind(mb)
	if err != nil {
		return err
	}

	log.Printf("Block: %d with hash %x mined by %s and received from Node %s\n", mb.Block.BH.BlockIndex, mb.Block.BH.BlockHash, mb.Miner, mb.Sender)

	o.Ch.Mu.Lock()
	defer o.Ch.Mu.Unlock()
	o.Ch.Chainstate.Loadchainstate()

	if mb.Block.BH.BlockIndex > o.Ch.LastBlock.BH.BlockIndex+2 {
		log.Println("Detecting a soft fork")
		return nil
	}

	if mb.Block.BH.BlockIndex-1 == o.Ch.LastBlock.BH.BlockIndex {
		// pause mining process that trying to mine this block
		o.Ch.Engine.Pause()
		log.Println("  Proccessing Block")

		if !BlockValidator(*mb.Block, o.Ch.Chainstate, o.Ch.LastBlock) {
			// resume paused mining process
			o.Ch.Engine.Resume()
			log.Printf("Block %x is not valid\n", mb.Block.BH.BlockHash)
			return fmt.Errorf("block %x is not valid", mb.Block.BH.BlockHash)

		}
		// stop mining process beacuse block mined by another node
		o.Ch.Engine.Close()
		fmt.Println()
		log.Printf("Block %x is valid\n", mb.Block.BH.BlockHash)

		o.Ch.MemPool.Chainstate.Utxos = o.Ch.Chainstate.Utxos
		o.Ch.LastBlock = *mb.Block
		o.Ch.ChainHeight++
		// Update NodeHeight of sender in KnownNodes
		o.Ch.KnownNodes[mb.Sender].NodeHeight++

		go o.Ch.AddBlockInDB(mb.Block, mb.Mu)

		// Broadcasting valid new Mined block in network
		// Reciver is BroadBlock function
		o.BroadChan <- mb

		o.Ch.SaveUtxoSet()

	}

	return nil
}

// save block sender in chainstate database
func SaveBlocksenderInDB(hash []byte, sender types.NodeID, d database.Database) error {

	err := d.DB.Put(hash, []byte(sender), nil)
	if err != nil {
		return err
	}
	log.Printf("Block: %x with sender %s saved in database\n", hash, sender)
	return nil
}
func (o *Objects) SendBlock(ctx echo.Context) error {
	gb := new(GetBlock)
	err := ctx.Bind(gb)
	if err != nil {
		return err
	}
	hash := gb.BlockHash

	block := core.ReadBlock(o.Ch.DB, hash)
	mb := NewMsgdBlock(block, types.NodeID(o.Ch.MinerAdd), block.BH.Miner)

	fmt.Printf("\nNode %s wants Block %x\n", gb.Node, block.BH.BlockHash)
	ctx.JSONPretty(200, mb, " ")

	return nil
}

func (o *Objects) SendInv(ctx echo.Context) error {
	inv := Inv{}
	inv.BlocksHash = make(map[blockIndex][]byte)
	inv.InvType = blockType
	iter := o.Ch.NewIter()
	for {
		block := iter.Next()
		if block.BH.PrevHash == nil {
			break
		}
		fmt.Printf("Adding block hash %x to inv\n", block.BH.BlockHash)
		inv.BlocksHash[blockIndex(block.BH.BlockIndex)] = block.BH.BlockHash
		inv.InvCount++
	}
	fmt.Printf("\nNode %s wants inventory \n", ctx.RealIP())

	ctx.JSONPretty(200, inv, " ")

	return nil

}

func (o *Objects) SendGen(ctx echo.Context) error {

	hash, err := o.Ch.DB.DB.Get([]byte("genesis_block"), nil)
	fmt.Printf("\nNode%s wants genesis block \n", ctx.RealIP())

	if err != nil {
		return err
	}

	block := core.ReadBlock(o.Ch.DB, hash)

	ctx.JSONPretty(200, *block, " ")

	return nil
}
