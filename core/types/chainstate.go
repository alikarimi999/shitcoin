package types

import (
	"bytes"
	"fmt"
	"log"
	"sync"

	"github.com/alikarimi999/shitcoin/database"
)

type ChainState struct {
	Mu    *sync.Mutex
	Utxos map[Account][]*UTXO
	DB    database.Database
}

func NewChainState() *ChainState {
	return &ChainState{
		Mu:    &sync.Mutex{},
		Utxos: make(map[Account][]*UTXO),
		DB:    database.Database{},
	}
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
				if bytes.Equal(in.OutPoint, utxo.Txid) && in.Vout == utxo.Index && in.Value == utxo.Txout.Value {
					fmt.Printf("One Token with %d Value deleted from %s UTXO Set in Pool UTXOSet\n ", utxo.Txout.Value, Pub2Address(utxo.Txout.PublicKeyHash, true))
					continue
				}
				utxos = append(utxos, utxo)

			}

		}
		u.Mu.Lock()
		u.Utxos[Account(Pub2Address(pk, false))] = utxos
		u.Mu.Unlock()

		utxos = []*UTXO{}

	}

	// add new Token
	var pkh []byte
	for index, out := range tx.TxOutputs {
		if out.Value == 0 {
			continue
		}
		pkh = out.PublicKeyHash
		utxo := &UTXO{tx.TxID, uint(index), out}
		u.Utxos[Account(Pub2Address(pkh, true))] = append(u.Utxos[Account(Pub2Address(pkh, true))], utxo)
		fmt.Printf("One Token with %d value added for %s in Pool UTXOSet from TX: %x\n", utxo.Txout.Value, Pub2Address(utxo.Txout.PublicKeyHash, true), tx.TxID)
	}

}

// this function read chainstate from chainstate databse and add it to in memory Chainstate
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

func (c *ChainState) Clean() {
	c.Mu = &sync.Mutex{}
	c.Utxos = make(map[Account][]*UTXO)
}

// validate block's transactions
// and if transaction is valid update in chainstate
func (Ch *ChainState) Validate_blk_trx(b Block) (map[Account][]*UTXO, bool) {

	for _, tx := range b.Transactions {
		if tx.IsCoinbase() {
			Ch.UpdateUtxoSet(tx)

			continue
		}
		trx := *tx
		if !Ch.Verifyhash(tx) || !trx.Checksig() {

			return nil, false
		}
		Ch.UpdateUtxoSet(tx)
	}

	return Ch.Utxos, true
}

// OP_EQUALVERIFY
func (u *ChainState) Verifyhash(tx *Transaction) bool {
	if !tx.IsCoinbase() {
		checker := []int{}

	IN:
		for _, in := range tx.TxInputs {
			pk := in.PublicKey
			var pkh []byte

			for _, utxo := range u.Utxos[Account(Pub2Address(pk, false))] {
				if bytes.Equal(in.OutPoint, utxo.Txid) && in.Vout == utxo.Index && in.Value == utxo.Txout.Value {
					pkh = utxo.Txout.PublicKeyHash
					if bytes.Equal(pkh, Hash160(pk)) {
						checker = append(checker, 1)
						continue IN
					}
				}
			}

		}

		if len(checker) == len(tx.TxInputs) {
			return true
		}
		return false
	}

	return true
}

func (c *ChainState) SnapShot() *ChainState {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	ch := &ChainState{
		Mu:    &sync.Mutex{},
		Utxos: make(map[Account][]*UTXO),
		DB:    c.DB,
	}

	for a, u := range c.Utxos {
		for _, utxo := range u {
			copy_utxo := *utxo
			ch.Utxos[a] = append(ch.Utxos[a], &copy_utxo)
		}
	}
	return ch
}
