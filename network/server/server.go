package server

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/alikarimi999/shitcoin/core"
	netype "github.com/alikarimi999/shitcoin/network/types"

	"github.com/labstack/echo/v4"
)

const (
	NETWORKPROTO string = "http://"
)

type Server struct {
	Mu           sync.Mutex
	Ch           *core.Chain
	Port         int
	RecievedTxs  [][]byte
	RecievedBlks [][]byte

	// channels
	TxCh  chan *netype.MsgTX
	BlkCh chan *netype.MsgBlock
}

func (s *Server) Run(wg *sync.WaitGroup) {

	defer wg.Done()
	e := echo.New()

	e.GET("/getutxo", s.sendUTXOs)
	e.POST("/sendtrx", s.getTrx)
	e.POST("/getinventory", s.SendInv)
	e.POST("/getblock", s.SendBlock)
	e.POST("/minedblock", s.MinedBlock)
	e.POST("/getnode", s.SendNodes)
	e.GET("nodeinfo", s.SendNodeInfo)

	e.GET("/peers", s.peers)
	e.GET("/block/:hash", s.block)
	e.GET("/height", s.SendHeight)

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", s.Port)))
}

// TODO: delete this and use node.NodeHeight
func (s *Server) SendHeight(ctx echo.Context) error {
	ctx.String(200, fmt.Sprintf("%d", atomic.LoadUint64(&s.Ch.ChainHeight)))
	return nil
}
