package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/alikarimi999/shitcoin/core"
	netype "github.com/alikarimi999/shitcoin/network/types"
)

// Broadcast received transaction to  Peers
func (c *Client) BroadTx(mt *netype.MsgTX) {

	c.PeerSet.Mu.Lock()
	defer c.PeerSet.Mu.Unlock()

	for _, n := range c.PeerSet.Peers {
		if mt.SenderID != n.ID {
			mt.SenderID = c.Ch.Node.ID
			b, _ := json.Marshal(mt)
			log.Printf("sending transaction %x to node %s\n", mt.TX.TxID, n.ID)
			c.Cl.Post(fmt.Sprintf("%s/sendtrx", n.FullAdd), "application/json", bytes.NewReader(b))
		}
	}
}

// this function download transactions from sync node transaction pool
func downloadTxPool(c *core.Chain, dst string) {
	inv, err := getInv(netype.TxType, c.Node.ID, c.Node.LastHash, dst, http.Client{Timeout: 20 * time.Second})
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if inv.InvType != netype.TxType {
		log.Println("incorrect data sended by sync node")
	}

	for _, tx := range inv.TXs {
		log.Printf("transaction %x recieved from %s\n", tx.TxID, inv.NodeId)
		if c.Validator.ValidateTX(tx) {
			c.TxPool.UpdatePool(tx, false)
			log.Printf("transaction %x is valid\n", tx.TxID)
			continue
		}
		log.Printf("transaction %x is not valid\n", tx.TxID)
	}

}
