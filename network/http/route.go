package network

import (
	"fmt"

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
	e.GET("/getver", o.SendVer)
	e.GET("/getgen", o.SendGen)
	e.POST("getdata", o.SendInv)
	e.POST("/getblock", o.SendBlock)
	e.POST("/minedblock", o.MinedBlock)
	e.POST("/intro", o.Intro)
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", port)))
}

func (o *Objects) Intro(ctx echo.Context) error {
	fmt.Println("Intro func")
	v := Version{}
	err := ctx.Bind(&v)
	if err != nil {
		return err
	}

	v.Address = fmt.Sprintf("http://%s:%s", ctx.RealIP(), v.Address)
	fmt.Println(v.Address)

	if !o.Ch.NodeExist(v.Address) {
		o.Ch.KnownNodes = append(o.Ch.KnownNodes, v.Address)
		fmt.Printf("Node %s added to known nodes list\n", v.Address)
	}
	return nil
}

func (o *Objects) MinedBlock(ctx echo.Context) error {
	mb := new(core.MinedBlock)
	err := ctx.Bind(mb)
	if err != nil {
		return err
	}
	fmt.Printf("Block %x mined by %s and received from Node %s\n", mb.Block.BH.BlockHash, mb.Miner, ctx.RealIP())

	if mb.Block.BH.BlockIndex-1 == o.Ch.LastBlock.BH.BlockIndex {
		fmt.Println("  Proccessing Block")
		if o.Ch.AddNewBlock(mb.Block) {
			o.Ch.LastBlock = mb.Block
		}

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

	fmt.Printf("\nNode %s wants Block %x\n", ctx.RealIP(), block.BH.BlockHash)
	ctx.JSONPretty(200, *block, " ")

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

func (o *Objects) SendVer(ctx echo.Context) error {
	v := &Version{fmt.Sprintf("%d", o.Port), o.Ch.LastBlock.BH.BlockHash, o.Ch.ChainHeight}
	fmt.Printf("\nNode %s wants version \n", ctx.RealIP())

	ctx.JSONPretty(200, v, " ")
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
