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
	Utxos   []*types.UTXO `json:"utxos"`
}

type MsgBlock struct {
	Mu     *sync.Mutex
	Sender string
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

type GetBlock struct {
	// node that request for block
	Node      string
	BlockHash []byte
}

type GetInv struct {
	NodeId   string
	InvType  InvType
	LastHash []byte
}

type Inv struct {
	NodeId   string
	InvCount int
	InvType  InvType

	// a map contain block index and block hash
	BlocksHash map[blockIndex][]byte

	// slice contain transation pool's transactions hash

	TXs []*types.Transaction
}

func NewInv() *Inv {
	return &Inv{
		NodeId:     "",
		InvCount:   0,
		InvType:    0,
		BlocksHash: make(map[blockIndex][]byte),
		TXs:        []*types.Transaction{},
	}

}

func Msgblock(b *types.Block, sender string, miner types.Address) *MsgBlock {
	mb := &MsgBlock{
		Sender: sender,
		Block:  b,
		Miner:  miner,
	}
	return mb
}
