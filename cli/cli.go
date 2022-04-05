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
	network "github.com/alikarimi999/shitcoin/network/http"
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

	port_new := newchain.Int("port", 0, "The port that node will listening on")
	miner_new := newchain.String("miner", "", "The Miner address")
	dbpath_new := newchain.String("dbpath", "", "Database Path")

	port_con := connect.Int("port", 0, "The port that node will listening on")
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
	c, err := core.NewChain(dbPath, port)
	if err != nil {
		log.Fatalln(err)
	}

	err = c.SetupChain(miner, 20)
	if err != nil {
		log.Fatalln(err)
	}
	c.MinerAdd = miner

	o := &network.Objects{
		Mu:        sync.Mutex{},
		Ch:        c,
		Port:      port,
		BroadChan: make(chan *network.MsgBlock),
		Cl:        http.Client{Timeout: 5 * time.Second},
	}

	go network.RunServer(o, port)

	var wg sync.WaitGroup
	wg.Add(1)
	go Miner(c, wg)

	wg.Wait()
}

func (cli *Commandline) Connect(miner []byte, node string, port int, dbPath string) {

	c := core.Loadchain(dbPath, port)
	c.MinerAdd = miner
	o := &network.Objects{
		Mu:        sync.Mutex{},
		Ch:        c,
		Port:      port,
		BroadChan: make(chan *network.MsgBlock),
		Cl:        http.Client{Timeout: 5 * time.Second},
	}

	go network.RunServer(o, port)

	err := network.PairNode(c, node)
	if err != nil {
		fmt.Println(err.Error())
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go o.IBD(wg)

	wg.Add(1)
	go Miner(c, wg)

	wg.Wait()

}

func Miner(c *core.Chain, wg sync.WaitGroup) {
	defer wg.Done()
	log.Println("Miner Function start!")

	for {

		// Sender is in AddTx2Pool Function
		b := <-c.BlockChann
		c.Mu.Lock()
		b.BH.BlockIndex = c.ChainHeight
		b.BH.PrevHash = c.LastBlock.BH.BlockHash
		c.Mu.Unlock()

		b.BH.Miner = c.MinerAdd

		if c.Engine.Start(b) {
			log.Printf("Block %d with hash %x with %d transations Mined successfully\n", b.BH.BlockIndex, b.BH.BlockHash, len(b.Transactions))

			// reciver is in BroadMinedBlock function
			c.MinedBlock <- b.SnapShot()

			c.Mu.Lock()
			c.ChainHeight++
			c.LastBlock = *b
			c.Mu.Unlock()

			err := b.SaveBlockInDB(&c.DB, &sync.Mutex{})
			if err != nil {
				log.Printf("Block %x did not add to database\n\n", b.BH.BlockHash)
			}
			log.Printf("Block %x successfully added to database\n\n", b.BH.BlockHash)

			// Now we have to add utxoset to database
			c.SaveUtxoSet()

		}
	}

}
