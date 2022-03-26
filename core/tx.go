package core

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"errors"
	"fmt"
	"math/big"
	"time"
)

type txsPool struct {
	Transactions []*Transaction
	Chainstate   *ChainState
}

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
		fmt.Println("transaction added")
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

// Transfer Transactions from transaction pool to Block
func (tp *txsPool) TransferTxs2Block(b *Block, amount int) error {

	tp.addMinerReward(b.BH.Miner, amount)

	b.Transactions = tp.Transactions

	// and make transaction pool Clean
	tp.Clean()

	return nil

}

func (tp *txsPool) addMinerReward(miner Address, amount int) error {
	pkh := Add2PKH(miner)
	reward := CoinbaseTx(pkh, amount)
	tp.Transactions = append(tp.Transactions, reward)

	tp.Chainstate.UpdateUtxoSet(reward)

	return nil
}

// Clean transaction pool
func (tp *txsPool) Clean() error {
	tp.Transactions = []*Transaction{}

	return nil
}

func (tx *Transaction) Print() {
	fmt.Printf("\n\tTX Hash: %x\n\n", tx.TxID)

	fmt.Printf("Inputs: \n\t")

	for i, in := range tx.TxInputs {
		fmt.Printf("\n Input: %d\n\t %x\n\tAccount: %s\tOutID: %x \n", i, in.OutPoint, Pub2Address(in.PublicKey, false), in.Vout)
	}
	fmt.Printf("\n\n")
	fmt.Printf("Outputs\n")

	for i, out := range tx.TxOutputs {
		fmt.Printf("Output: %d\tAccount: %s\tValue: %v\n", i, Pub2Address(out.PublicKeyHash, true), out.Value)

	}

}
