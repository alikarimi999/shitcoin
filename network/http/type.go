package network

import (
	"github.com/alikarimi999/shitcoin/core"
)

type InvType int
type blockIndex uint64

const (
	blockType InvType = iota
	txType
)

type msgUTXOSet struct {
	Account core.Account `json:"account"`
	Utxos   []core.UTXO  `json:"utxos"`
}

type NodeInfo struct {
	Sender core.NodeID
	NodeId core.NodeID
	// node FullAdd
	Port       string
	FullAdd    string
	LastHash   []byte
	NodeHeight uint64
}

type GetData struct {
	// last block hash that nodes hash
	// nil mean does not have even genesis block
	// sync node return inv struct which contain hash of all block that it's have
	LastHash []byte
}

type GetBlock struct {
	BlockHash []byte
}

type GetTX struct {
	TxHash []byte
}

type Inv struct {
	InvCount int
	InvType  InvType

	// a map contain block index and block hash
	BlocksHash map[blockIndex][]byte
}
