package network

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/alikarimi999/shitcoin/core"
	"github.com/labstack/echo/v4"
)

type Objects struct {
	Ch   *core.Chain
	Port int
}

func RunServer(c *core.Chain, port int) {

	o := Objects{c, port}
	go o.Ch.Miner()
	e := echo.New()
	e.GET("/getutxo", o.sendUtxoset)
	e.POST("sendtrx", o.getTrx)
	e.GET("/getnodeinfo", o.SendNodeInfo)
	e.GET("/getgen", o.SendGen)
	e.POST("getdata", o.SendInv)
	e.POST("/getblock", o.SendBlock)
	e.POST("/minedblock", o.MinedBlock)
	e.POST("/newnode", o.NewNode)
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", port)))
}

func (o *Objects) NewNode(ctx echo.Context) error {

	n := &NodeInfo{}
	err := ctx.Bind(n)
	if err != nil {
		return err
	}

	if n.FullAdd == "" && n.Sender == n.NodeId {
		n.FullAdd = fmt.Sprintf("http://%s:%s", ctx.RealIP(), n.Port)
	}
	// Add to known nodes
	n.AddNode(o.Ch)

	n.BroadNode(o.Ch, http.Client{Timeout: 5 * time.Second})
	return nil
}

func (o *Objects) MinedBlock(ctx echo.Context) error {
	mb := new(core.MsgBlock)
	err := ctx.Bind(mb)
	if err != nil {
		return err
	}
	fmt.Printf("Block %x mined by %s and received from Node %s\n", mb.Block.BH.BlockHash, mb.Miner, mb.Sender)
	fmt.Println(mb.Block.BH.BlockIndex, o.Ch.LastBlock.BH.BlockIndex)
	if mb.Block.BH.BlockIndex-1 == o.Ch.LastBlock.BH.BlockIndex {
		fmt.Println("  Proccessing Block")
		if !o.Ch.BlockValidator(*mb.Block) {
			fmt.Printf("Block %x is not valid\n", mb.Block.BH.BlockHash)
			return errors.New(fmt.Sprintf("block %x is not valid\n", mb.Block.BH.BlockHash))

		}
		fmt.Printf("Block %x is valid\n", mb.Block.BH.BlockHash)

		// Broadcasting valid new Mined block in network
		o.Ch.BroadBlock(mb, http.Client{Timeout: 5 * time.Second})

		o.Ch.AddBlockInDB(mb.Block)
		o.Ch.SyncUtxoSet()

		o.Ch.LastBlock = mb.Block

	}

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
	mb := core.NewMsgdBlock(block, core.NodeID(o.Ch.MinerAdd), block.BH.Miner)

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

func (o *Objects) SendNodeInfo(ctx echo.Context) error {
	n := NewNodeInfo(o.Ch, o.Port)

	fmt.Printf("\nNode %s wants version \n", ctx.RealIP())

	ctx.JSONPretty(200, n, " ")
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
