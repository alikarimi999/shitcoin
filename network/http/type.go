package network

import (
	"sync"

	"github.com/alikarimi999/shitcoin/core/types"
)

type InvType int
type blockIndex uint64

const (
	blockType InvType = iota
	txType
)

type msgUTXOSet struct {
	Account types.Account `json:"account"`
	Utxos   []types.UTXO  `json:"utxos"`
}

type MsgBlock struct {
	Mu     *sync.Mutex
	Sender types.NodeID
	Block  *types.Block
	Miner  types.Address
}

func NewMsgBlock() *MsgBlock {
	m := &MsgBlock{
		Mu:    new(sync.Mutex),
		Block: types.NewBlock(),
	}
	return m
}

type GetNode struct {
	SrcNodes   []*types.Node
	ShareNodes []*types.Node
}

type GetData struct {
	// last block hash that nodes has
	// nil mean does not have even genesis block
	// sync node return inv struct which contain hash of all block that it's have
	LastHash []byte
}

type GetBlock struct {
	// node that request for block
	Node      types.NodeID
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

func NewMsgdBlock(b *types.Block, sender types.NodeID, miner types.Address) *MsgBlock {
	mb := &MsgBlock{
		Sender: sender,
		Block:  b,
		Miner:  miner,
	}
	return mb
}
