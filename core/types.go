package core

import (
	"log"

	"github.com/alikarimi999/shitcoin/database"
)

type Chainid int

type publickey []byte

type txid []byte

// account address in byte
type Address []byte

// Account address in string
type Account string

type UTXO struct {
	Txid  []byte
	Index uint
	Txout *TxOut
}

type ChainState struct {
	Utxos map[Account][]*UTXO
	DB    database.Database
}

func FailOnErro(err error, s string) {
	if err != nil {
		log.Fatalf("%s:\t%s", err.Error(), s)
	}
}
