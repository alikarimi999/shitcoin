package types

import (
	"sync"

	"github.com/alikarimi999/shitcoin/core/types"
)

type InvType int
type BlockIndex uint64

const (
	BlockType InvType = iota
	TxType
)

type MsgUTXOSet struct {
	Account types.Account `json:"account"`
	Utxos   []*types.UTXO `json:"utxos"`
}

type MsgTX struct {
	SenderID string            `json:"sender"`
	TX       types.Transaction `json:"tx"`
}

type MsgBlock struct {
	Mu     *sync.Mutex
	Sender string
	Block  *types.Block
	Miner  string
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
	BlocksHash map[BlockIndex][]byte

	// slice contain transation pool's transactions hash

	TXs []*types.Transaction
}

func NewInv() *Inv {
	return &Inv{
		NodeId:     "",
		InvCount:   0,
		InvType:    0,
		BlocksHash: make(map[BlockIndex][]byte),
		TXs:        []*types.Transaction{},
	}

}

func Msgblock(b *types.Block, sender string, miner string) *MsgBlock {
	mb := &MsgBlock{
		Sender: sender,
		Block:  b,
		Miner:  miner,
	}
	return mb
}