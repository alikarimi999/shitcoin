package core

import (
	"bytes"
	"errors"
	"fmt"
	"log"

	"github.com/alikarimi999/shitcoin/core/types"
)

// adding transaction to transaction Pool
func (c *Chain) AddTx2Pool(tx *types.Transaction) error {

	c.MemPool.Mu.Lock()
	defer c.MemPool.Mu.Unlock()

	// these are transactions that added to block and sended to Miner function
	// and deleted from mempool transactions
	in_block_txs := []*types.Transaction{}

	for _, t := range in_block_txs {
		if bytes.Equal(t.TxID, tx.TxID) {
			return fmt.Errorf("transaction %x exist in block that is mining now", tx.TxID)
		}
	}

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

				in_block_txs = c.MemPool.Transactions
				c.MemPool.Clean()
				// Reciver is in Miner function
				c.BlockChann <- b
				log.Println("Sending block to Mine function through BlockChann")
			}
		}
		return nil
	}
	return errors.New("Transaction is not Valid")
}
