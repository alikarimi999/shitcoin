package core

import "github.com/alikarimi999/shitcoin/database"

type ChainState struct {
	Utxos map[Account][]*UTXO
	DB    database.Database
}

// validate block's transactions
// and if transaction is valid update in memory UTXO set
func (ch *ChainState) Validate_blk_trx(b Block) (map[Account][]*UTXO, bool) {

	tempChainstate := &ChainState{}
	tempChainstate.Utxos = make(map[Account][]*UTXO)
	tempChainstate.Utxos = ch.Utxos

	for _, tx := range b.Transactions {

		if tx.IsCoinbase() {
			tempChainstate.UpdateUtxoSet(tx)

			continue
		}
		trx := *tx
		if !tempChainstate.Verifyhash(tx) || !trx.Checksig() {

			return nil, false
		}
		tempChainstate.UpdateUtxoSet(tx)
	}

	return tempChainstate.Utxos, true
}

func (c *ChainState) SnapShot() *ChainState {
	ch := &ChainState{
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
