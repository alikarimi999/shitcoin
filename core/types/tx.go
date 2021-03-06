package types

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"log"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil"
)

type Transaction struct {
	Timestamp time.Duration
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
	PublicKey []byte // serialized compressed public key
	Signature []byte
}

func (in *TxIn) serialize() []byte {
	b := bytes.Join(
		[][]byte{
			in.OutPoint,
			Int2Hex(int64(in.Vout)),
			Int2Hex(int64(in.Value)),
			in.PublicKey,
			in.Signature,
		}, nil,
	)
	return b
}

func (out *TxOut) serialize() []byte {
	b := bytes.Join(
		[][]byte{
			out.PublicKeyHash,
			Int2Hex(int64(out.Value)),
		}, nil,
	)
	return b
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
					if bytes.Equal(pkh, btcutil.Hash160(in.PublicKey)) {
						checker = append(checker, 1)
						continue IN
					}
				}
			}

		}
		if len(checker) == len(tx.TxInputs) && tx.Checksig() && tx.CheckHash() {
			return true
		}
		return false
	}

	return tx.CheckHash()
}

func (tx *Transaction) CheckHash() bool {
	snapTx := tx.SnapShot()

	hash := snapTx.TxID
	snapTx.TxID = nil
	for _, in := range snapTx.TxInputs {
		in.PublicKey = nil
		in.Signature = nil
	}
	if !bytes.Equal(hash, Hash(snapTx)) {
		log.Println("Transaction hash incorrect")
		return false
	}
	return true
}

func (tx *Transaction) serialize() []byte {

	b := bytes.Join(
		[][]byte{
			Int2Hex(int64(tx.Timestamp)),
			tx.TxID,
			join(tx.TxInputs),
			join(tx.TxOutputs),
		},
		nil,
	)

	return b

}

// TODO: for sign interface
func Hash(s serializer) []byte {

	data := s.serialize()

	hash := sha256.Sum256(data)

	return hash[:]
}

func CoinbaseTx(to Address, amount int) *Transaction {
	tx := &Transaction{
		Timestamp: time.Duration(time.Now().UTC().UnixNano()),
		TxID:      []byte{},
		TxInputs:  []*TxIn{},
		TxOutputs: []*TxOut{
			{
				PublicKeyHash: to,
				Value:         amount,
			},
		},
	}

	tx.TxID = Hash(tx)
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

	//TODO: delete this
	// txCopy := tx.TrimmeTX()

	// curve := elliptic.P256()
	for _, in := range tx.TxInputs {

		sig, err := btcec.ParseDERSignature(in.Signature, btcec.S256())
		if err != nil {
			fmt.Println(err.Error())
			return false
		}
		pubKey, err := btcec.ParsePubKey(in.PublicKey, btcec.S256())
		if err != nil {
			fmt.Println(err.Error())
			return false
		}

		if !sig.Verify(tx.TxID, pubKey) {
			log.Println("Signature does not match")
			return false
		}

		//TODO: delete this
		// 	x := big.Int{}
		// 	y := big.Int{}

		// 	keyLen := len(pubKey)

		// 	x.SetBytes(pubKey[:(keyLen / 2)])
		// 	y.SetBytes(pubKey[(keyLen / 2):])

		// 	rawpubkey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}

		// 	// convert signature to r,s
		// 	r := big.Int{}
		// 	s := big.Int{}

		// 	sigLen := len(sig)

		// 	s.SetBytes(sig[(sigLen / 2):])
		// 	r.SetBytes(sig[:(sigLen / 2)])

		// 	if !ecdsa.Verify(&rawpubkey, txCopy.serialize(), &r, &s) {
		// 		log.Println("Signature does not match")
		// 		return false
		// 	}

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
