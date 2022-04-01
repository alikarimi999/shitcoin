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

type MsgBlock struct {
	Sender core.NodeID
	Block  *core.Block
	Miner  core.Address
}

type GetNode struct {
	SrcNodes   []*core.Node
	ShareNodes []*core.Node
}

type GetData struct {
	// last block hash that nodes has
	// nil mean does not have even genesis block
	// sync node return inv struct which contain hash of all block that it's have
	LastHash []byte
}

type GetBlock struct {
	// node that request for block
	Node      core.NodeID
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

func NewMsgdBlock(b *core.Block, sender core.NodeID, miner core.Address) *MsgBlock {
	mb := &MsgBlock{
		Sender: sender,
		Block:  b,
		Miner:  miner,
	}
	return mb
}
