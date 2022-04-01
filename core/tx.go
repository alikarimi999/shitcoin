package core

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"errors"
	"fmt"
	"log"
	"math/big"
	"time"
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
func (c *Chain) AddTx2Pool(tx *Transaction) error {

	for _, t := range c.MemPool.Transactions {
		if bytes.Equal(t.TxID, tx.TxID) {
			return fmt.Errorf("transaction %x exist in %s Mem Pool", tx.TxID, NodeID(c.MinerAdd))
		}
	}

	trx := *tx

	if c.MemPool.Chainstate.Verifyhash(tx) && trx.Checksig() {
		c.MemPool.Transactions = append(c.MemPool.Transactions, tx)

		// We have to update mem Pool Utxo set if we add transaction to Poll

		c.MemPool.Chainstate.UpdateUtxoSet(tx)
		log.Println("transaction added Mem Pool")
		if len(c.MemPool.Transactions) >= BlockMaxTransactions-1 {

			b := NewBlock()
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

// OP_EQUALVERIFY
func (u *ChainState) Verifyhash(tx *Transaction) bool {
	if !tx.IsCoinbase() {
		for _, in := range tx.TxInputs {
			pk := in.PublicKey
			var pkh []byte
			for _, utxo := range u.Utxos[Account(Pub2Address(pk, false))] {
				if in.Vout == utxo.Index {
					pkh = utxo.Txout.PublicKeyHash
					break
				}
			}

			if !bytes.Equal(pkh, Hash160(pk)) {
				return false
			}
		}
	}

	return true
}

// OP_CHECKSIG
func (tx Transaction) Checksig() bool {
	if tx.IsCoinbase() {
		return true
	}

	txCopy := tx.TrimmeTX()

	curve := elliptic.P256()
	for _, in := range tx.TxInputs {

		sig := in.Signature
		pubKey := in.PublicKey

		x := big.Int{}
		y := big.Int{}

		keyLen := len(pubKey)

		x.SetBytes(pubKey[:(keyLen / 2)])
		y.SetBytes(pubKey[(keyLen / 2):])

		rawpubkey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}

		// convert signature to r,s
		r := big.Int{}
		s := big.Int{}

		sigLen := len(sig)

		s.SetBytes(sig[(sigLen / 2):])
		r.SetBytes(sig[:(sigLen / 2)])

		if !ecdsa.Verify(&rawpubkey, txCopy.Serialize(), &r, &s) {
			fmt.Println("Signature does not match")
			return false
		}

	}
	return true
}

func (tx Transaction) TrimmeTX() *Transaction {

	var inputs []*TxIn
	var outputs []*TxOut

	for _, in := range tx.TxInputs {
		inputs = append(inputs, &TxIn{in.OutPoint, in.Vout, in.Value, nil, nil})
	}

	for _, out := range tx.TxOutputs {
		outputs = append(outputs, &TxOut{out.PublicKeyHash, out.Value})
	}

	txCopy := &Transaction{tx.Timestamp, tx.TxID, inputs, outputs}

	return txCopy
}

func (tx *Transaction) SnapShot() *Transaction {

	trx := &Transaction{}

	for _, in := range tx.TxInputs {
		copy_in := *in
		trx.TxInputs = append(trx.TxInputs, &copy_in)
	}

	for _, out := range tx.TxOutputs {
		copy_out := *out
		trx.TxOutputs = append(trx.TxOutputs, &copy_out)
	}
	return trx
}
