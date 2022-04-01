package core

import (
	"fmt"
	"log"
)

type UTXO struct {
	Txid  []byte
	Index uint
	Txout *TxOut
}

func (c *Chain) SyncUtxoSet() error {

	saveUTXOsInDB(*c.MemPool.Chainstate)

	return nil
}

// when node receive a transaction and verify it
// then delete Tokens that used in inputs of transaction from it's utxo set
// and add new Token of transaction to it's utxoset
func (u *ChainState) UpdateUtxoSet(tx *Transaction) {

	utxos := []*UTXO{}
	// delete spent Token
	if !tx.IsCoinbase() {

		pk := tx.TxInputs[0].PublicKey
		for _, in := range tx.TxInputs {

			for _, utxo := range u.Utxos[Account(Pub2Address(pk, false))] {
				if in.Vout == utxo.Index {
					fmt.Printf("One Token with %d Value deleted from %s UTXO Set in Pool UTXOSet\n ", utxo.Txout.Value, Pub2Address(utxo.Txout.PublicKeyHash, true))
					continue
				}
				utxos = append(utxos, utxo)

			}

		}
		u.Utxos[Account(Pub2Address(pk, false))] = utxos

		utxos = []*UTXO{}

	}

	// add new Token
	var pkh []byte
	for index, out := range tx.TxOutputs {
		if out.Value == 0 {
			break
		}
		pkh = out.PublicKeyHash
		utxo := &UTXO{tx.TxID, uint(index), out}
		u.Utxos[Account(Pub2Address(pkh, true))] = append(u.Utxos[Account(Pub2Address(pkh, true))], utxo)
		fmt.Printf("One Token with %d value added for %s in Pool UTXOSet\n", utxo.Txout.Value, Pub2Address(utxo.Txout.PublicKeyHash, true))
	}

}

// this method read UTXOs from blockchain database and add it to in memory chain state
func (c *Chain) Retrieve_chainstate_on_db() {

	spentTxouts := make(map[string][]int)
	iter := c.NewIter()

	for {
		block := iter.Next()
		fmt.Printf("Proccessing Block: %x from Database\n", block.BH.BlockHash)
		for j := len(block.Transactions) - 1; j >= 0; j-- {
			tx := block.Transactions[j]
		Output:
			for outindex, out := range tx.TxOutputs {
				if spentTxouts[string(tx.TxID)] != nil {
					for _, spentIndex := range spentTxouts[string(tx.TxID)] {
						if spentIndex == outindex {
							continue Output
						}
					}
				}

				utxo := &UTXO{tx.TxID, uint(outindex), out}
				c.Chainstate.Utxos[Account(Pub2Address(out.PublicKeyHash, true))] = append(c.Chainstate.Utxos[Account(Pub2Address(out.PublicKeyHash, true))], utxo)

				fmt.Printf("Token with value %d added to UTXO set of %s\n\n", utxo.Txout.Value, Pub2Address(utxo.Txout.PublicKeyHash, true))
			}
			if !tx.IsCoinbase() {
				for _, in := range tx.TxInputs {
					spentTxouts[string(in.OutPoint)] = append(spentTxouts[string(in.OutPoint)], int(in.Vout))
				}
			}

		}
		if len(block.BH.PrevHash) == 0 {
			break
		}
	}
	c.MemPool.Chainstate.Utxos = c.Chainstate.Utxos
	c.Chainstate.Utxos = make(map[Account][]*UTXO)

}

// this function read chainstate from chainstate databse and add it to in memory Chainstate in MemPool
func (ch *ChainState) Loadchainstate() {

	utxos := make(map[Account][]*UTXO)
	iter := ch.DB.DB.NewIterator(nil, nil)
	for iter.Next() {

		key := iter.Key()
		value := iter.Value()
		u := []*UTXO{}
		ut := Deserialize(value, &u)
		utxo, ok := ut.(*[]*UTXO)
		if ok {
			utxos[Account(key)] = append(utxos[Account(key)], *utxo...)
			for _, utx := range *utxo {
				fmt.Printf("One token with %d value added to %s UTXO set\n\n", utx.Txout.Value, Pub2Address(utx.Txout.PublicKeyHash, true))

			}
		} else {
			fmt.Println("Data Corupted")
			continue
		}
	}

	iter.Release()
	err := iter.Error()
	if err != nil {
		log.Fatalln(err)
	}
	ch.Utxos = utxos
}
