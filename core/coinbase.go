package core

import "time"

func CoinbaseTx(to Address, amount int) *Transaction {
	tx := &Transaction{
		Timestamp: time.Now(),
		TxID:      []byte{},
		TxInputs:  []*TxIn{},
		TxOutputs: []*TxOut{
			{
				PublicKeyHash: to,
				Value:         amount,
			},
		},
	}

	tx.SetHash()
	return tx
}

func (tx *Transaction) IsCoinbase() bool {
	return len(tx.TxInputs) == 0
}
