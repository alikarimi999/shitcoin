package network

import (
	"fmt"
	"sync"

	"github.com/alikarimi999/shitcoin/core"
	"github.com/labstack/echo/v4"
)

const (
	NETWORKPROTO string = "http://"
)

type Server struct {
	Mu           sync.Mutex
	Ch           *core.Chain
	Port         int
	Client       Client
	RecievedTxs  [][]byte
	RecievedBlks [][]byte
}

func RunServer(s *Server, port int) {

	go s.Client.BroadMinedBlock()
	go s.Client.BroadBlock()

	e := echo.New()

	e.GET("/getutxo", s.sendUTXOs)
	e.POST("/sendtrx", s.getTrx)
	e.POST("/getinventory", s.SendInv)
	e.POST("/getblock", s.SendBlock)
	e.POST("/minedblock", s.MinedBlock)
	e.POST("/getnode", s.SendNodes)
	e.GET("nodeinfo", s.SendNodeInfo)
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", port)))
}
