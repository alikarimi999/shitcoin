package types

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
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

func (tx *Transaction) IsValid(u []*UTXO) bool {
	if !tx.IsCoinbase() {
		checker := []int{}

	IN:
		for _, in := range tx.TxInputs {
			var pkh []byte

			for _, utxo := range u {
				if bytes.Equal(in.OutPoint, utxo.Txid) && in.Vout == utxo.Index && in.Value == utxo.Txout.Value {
					pkh = utxo.Txout.PublicKeyHash
					if bytes.Equal(pkh, Hash160(in.PublicKey)) {
						checker = append(checker, 1)
						continue IN
					}
				}
			}

		}

		if len(checker) == len(tx.TxInputs) && tx.Checksig() {
			return true
		}
		return false
	}

	return tx.Checksig()
}

func SerializeTxs(txs []*Transaction) []byte {

	var result []byte
	for _, tx := range txs {
		result = append(result, tx.Serialize()...)
	}

	return result
}

func (tx *Transaction) Serialize() []byte {

	buff := new(bytes.Buffer)

	encoder := gob.NewEncoder(buff)
	encoder.Encode(tx)

	return buff.Bytes()

}

func (tx *Transaction) SetHash() {
	for _, in := range tx.TxInputs {
		in.Signature = nil
	}
	data := tx.Serialize()

	hash := sha256.Sum256(data)

	tx.TxID = hash[:]
}

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

// deep copy of transaction
func (tx *Transaction) SnapShot() *Transaction {

	trx := &Transaction{
		Timestamp: tx.Timestamp,
		TxID:      tx.TxID,
	}

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
