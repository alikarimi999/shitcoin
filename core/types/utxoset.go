package types

import (
	"bytes"
	"sync"
)

type UtxoSet struct {
	Mu     *sync.Mutex
	Tokens map[Account][]*UTXO
}

func NewUtxoSet() *UtxoSet {
	return &UtxoSet{
		Mu:     &sync.Mutex{},
		Tokens: make(map[Account][]*UTXO),
	}
}

func (u *UtxoSet) SnapShot() *UtxoSet {
	u.Mu.Lock()
	defer u.Mu.Unlock()
	us := &UtxoSet{
		Mu:     &sync.Mutex{},
		Tokens: make(map[Account][]*UTXO),
	}

	for a, utxos := range u.Tokens {
		for _, ut := range utxos {
			us.Tokens[a] = append(us.Tokens[a], ut.SnapShot())
		}
	}

	return us
}

func (u *UtxoSet) UpdateUtxoSet(tx *Transaction) {

	u.Mu.Lock()
	defer u.Mu.Unlock()
	// delete spent Token
	if !tx.IsCoinbase() {

		pk := tx.TxInputs[0].PublicKey
		account := Account(PK2Add(pk, false))
		for _, in := range tx.TxInputs {

			for i, utxo := range u.Tokens[account] {
				if bytes.Equal(in.OutPoint, utxo.Txid) && in.Vout == utxo.Index && in.Value == utxo.Txout.Value {
					u.Tokens[account] = append(u.Tokens[account][:i], u.Tokens[account][i+1:]...)
					// fmt.Printf("One Token with %d Value deleted from %s\n ", utxo.Txout.Value, Pub2Address(utxo.Txout.PublicKeyHash, true))
					continue
				}
			}

		}

	}

	// add new Token
	var pkh []byte
	for index, out := range tx.TxOutputs {
		if out.Value == 0 {
			continue
		}
		pkh = out.PublicKeyHash
		utxo := &UTXO{
			Txid:  tx.TxID,
			Index: uint(index),
			Txout: out,
		}
		u.Tokens[Account(PK2Add(pkh, true))] = append(u.Tokens[Account(PK2Add(pkh, true))], utxo)
		// fmt.Printf("One Token with %d value added for %s in UTXO Set\n", utxo.Txout.Value, Pub2Address(utxo.Txout.PublicKeyHash, true))
	}

}
