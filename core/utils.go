package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"

	"github.com/mr-tron/base58"
	"golang.org/x/crypto/ripemd160"
)

func Add2PKH(a Address) []byte {
	publickeyhash, err := base58.Decode(string(a))
	if err != nil {
		log.Fatal(err)
	}
	publickeyhash = publickeyhash[1 : len(publickeyhash)-4]

	return publickeyhash
}

func PK2Add(pk []byte) Account {
	publicKeyHash := Hash160(pk)
	versionedHahs := AddVersion(publicKeyHash, byte(0x00))
	check := Checksum(versionedHahs)

	cHash := append(versionedHahs, check...)

	address := base58.Encode(cHash)

	return Account(address)

}

func Checksum(b []byte) []byte {
	h1 := sha256.Sum256(b)
	h2 := sha256.Sum256(h1[:])

	return h2[:4]
}

// this function add addressVersion
func AddVersion(b []byte, v byte) []byte {

	return append([]byte{v}, b...)
}

// this is obvoious
func Hash160(pub []byte) []byte {
	hash := sha256.Sum256(pub)

	hasher := ripemd160.New()
	_, err := hasher.Write(hash[:])
	if err != nil {
		log.Panic(err)
	}

	pkh := hasher.Sum(nil)

	return pkh
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

// Conver Pub Key to Coin address

func Pub2Address(pub []byte, hash bool) []byte {

	if !hash {
		pub = Hash160(pub)
	}
	versionedHahs := AddVersion(pub, byte(0x00))
	check := Checksum(versionedHahs)

	cHash := append(versionedHahs, check...)

	address := base58.Encode(cHash)

	return []byte(address)
}
