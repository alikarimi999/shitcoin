package cli

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/alikarimi999/shitcoin/core"
	"github.com/alikarimi999/shitcoin/network"
	"github.com/alikarimi999/shitcoin/network/client"
	"github.com/alikarimi999/shitcoin/network/server"
	netype "github.com/alikarimi999/shitcoin/network/types"
)

type Commandline struct{}

func (cli *Commandline) ValidateArgs() {
	if len(os.Args) < 2 {
		cli.Printusage()
		runtime.Goexit()
	}
}

func (cli *Commandline) Printusage() {
	fmt.Println("Usage:")
	fmt.Println(" newchain - newchain will start a newchain from scratch and creat a genesis block")
	fmt.Println(" connect - connect start a new node and connect it to other nodes on network")

}

func (cli *Commandline) Run() {
	cli.ValidateArgs()

	newchain := flag.NewFlagSet("newchain", flag.ExitOnError)
	connect := flag.NewFlagSet("connect", flag.ExitOnError)

	port_new := newchain.Int("port", 5000, "The port that node will listening on")
	miner_new := newchain.String("miner", "", "The Miner address")
	dbpath_new := newchain.String("dbpath", "", "Database Path")

	port_con := connect.Int("port", 5000, "The port that node will listening on")
	miner_con := connect.String("miner", "", "The Miner address")
	node_address := connect.String("address", "", "The node address that we want to connect for firsttime")
	dbpath_con := connect.String("dbpath", "", "Database Path")

	switch os.Args[1] {
	case "newchain":
		err := newchain.Parse(os.Args[2:])
		if err != nil {
			log.Fatalln(err)
		}
	case "connect":
		err := connect.Parse(os.Args[2:])
		if err != nil {
			log.Fatalln(err)
		}

	default:
		cli.Printusage()
	}

	if newchain.Parsed() {
		if *port_new == 0 || *miner_new == "" || *dbpath_new == "" {
			newchain.Usage()
			runtime.Goexit()
		}
		cli.NewChain([]byte(*miner_new), *port_new, *dbpath_new)
	}

	if connect.Parsed() {
		if *port_con == 0 || *miner_con == "" || *node_address == "" || *dbpath_con == "" {
			connect.Usage()
			runtime.Goexit()
		}
		cli.Connect([]byte(*miner_con), *node_address, *port_con, *dbpath_con)
	}
}

func (cli *Commandline) NewChain(miner []byte, port int, dbPath string) {
	c, err := core.NewChain(dbPath, port, miner)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Starting Node %s\n", c.Node.ID)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go c.ChainState.Handler(wg)
	wg.Add(1)
	go c.TxPool.Handler(wg)
	wg.Add(1)
	go c.Miner.Handler(wg)

	err = c.SetupChain()
	if err != nil {
		log.Fatalln(err)
	}
	client := &client.Client{
		Ch:      c,
		PeerSet: netype.NewPeerSet(),
		Cl:      http.Client{Timeout: 20 * time.Second},
	}
	server := &server.Server{
		Mu:           sync.Mutex{},
		Ch:           c,
		Port:         port,
		RecievedTxs:  make([][]byte, 30, 60),
		RecievedBlks: make([][]byte, 10, 20),

		PeerSet: client.PeerSet,
		TxCh:    make(chan *netype.MsgTX),
		BlkCh:   make(chan *netype.MsgBlock),
	}

	wg.Add(1)
	go network.Setup(client, server, c, wg)

	wg.Wait()
}

func (cli *Commandline) Connect(miner []byte, node string, port int, dbPath string) {

	c, err := core.Loadchain(dbPath, port, miner)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("starting node %s\n", c.Node.ID)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go c.ChainState.Handler(wg)
	wg.Add(1)
	go c.TxPool.Handler(wg)
	wg.Add(1)
	go c.Miner.Handler(wg)

	client := &client.Client{
		Ch:      c,
		Cl:      http.Client{Timeout: 20 * time.Second},
		PeerSet: netype.NewPeerSet(),
	}

	s := &server.Server{
		Mu:           sync.Mutex{},
		Ch:           c,
		Port:         port,
		RecievedTxs:  make([][]byte, 30, 60),
		RecievedBlks: make([][]byte, 10, 20),
		PeerSet:      client.PeerSet,

		TxCh:  make(chan *netype.MsgTX),
		BlkCh: make(chan *netype.MsgBlock),
	}

	err = client.Peers(node)
	if err != nil {
		fmt.Println(err.Error())
	}
	client.IBD()

	wg.Add(1)
	go network.Setup(client, s, c, wg)

	wg.Wait()

}
