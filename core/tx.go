package core

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/alikarimi999/shitcoin/core/types"
)

type Transaction struct {
	Timestamp time.Time
	TxID      []byte
	TxInputs  []*TxIn
	TxOutputs []*TxOut
}

type TxOut struct {
	PublicKeyHash []byte
	Value         int
}

type TxIn struct {
	OutPoint  []byte
	Vout      uint
	Value     int
	PublicKey []byte
	Signature []byte
}

// adding transaction to transaction Pool
func (c *Chain) AddTx2Pool(tx *types.Transaction) error {

	for _, t := range c.MemPool.Transactions {
		if bytes.Equal(t.TxID, tx.TxID) {
			return fmt.Errorf("transaction %x exist in %s Mem Pool", tx.TxID, types.NodeID(c.MinerAdd))
		}
	}

	trx := *tx

	if c.MemPool.Chainstate.Verifyhash(tx) && trx.Checksig() {
		c.MemPool.Transactions = append(c.MemPool.Transactions, tx)

		// We have to update mem Pool Utxo set if we add transaction to Poll

		c.MemPool.Chainstate.UpdateUtxoSet(tx)
		log.Println("transaction added Mem Pool")
		if len(c.MemPool.Transactions) >= BlockMaxTransactions-1 {

			b := types.NewBlock()
			if err := c.MemPool.TransferTxs2Block(b, c.MinerAdd, 15); err != nil {
				log.Println(err.Error())
			} else {
				log.Println("Mem Pool has 3 Transaction")
				// Reciver is in Miner function
				c.BlockChann <- b
				log.Println("Sending block to Mine function through BlockChann")
			}
		}
		return nil
	}
	return errors.New("Transaction is not Valid")
}
