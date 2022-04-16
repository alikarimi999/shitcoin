package core

import (
	"bytes"

	"github.com/alikarimi999/shitcoin/core/types"
	"github.com/alikarimi999/shitcoin/database"
)

type IterType int

const (
	minerReward int      = 15
	IterByHash  IterType = iota
	IterByIndex
)

type BlkIterator struct {
	iterType IterType // iterate base on hash or index
	// start point can be hash or index of newer mined block
	start interface{}
	// start point can be hash or index
	// nil means iterate until reach the genesis block
	stop interface{}
	db   database.DB
	// TODO: implement reverse
	// by defautl iterate from newer block to older one
	reverse bool
}

func NewIter(iterType IterType, start, stop interface{}, db database.DB) *BlkIterator {

	return &BlkIterator{
		iterType: iterType,
		start:    start,
		stop:     stop,
		db:       db,
		reverse:  false,
	}
}

func (i *BlkIterator) Next() (*types.Block, error) {

	if i.iterType == IterByIndex {
		startHash, err := i.db.GetBlkHash(i.start.(uint64), nil)
		if err != nil {
			return nil, err
		}
		stopHash, err := i.db.GetBlkHash(i.stop.(uint64), nil)
		if err != nil {
			return nil, err
		}
		i.iterType = IterByHash
		i.start = startHash
		i.stop = stopHash
	}

	b, err := i.db.GetBlockH(i.start.([]byte), nil)
	if err != nil {
		return nil, err
	}
	i.start = b.BH.PrevHash

	return b, nil
}

func (i *BlkIterator) Run() ([]*types.Block, error) {
	result := []*types.Block{}
	for {
		b, err := i.Next()
		if err != nil {
			return nil, err
		}
		result = append(result, b)
		if bytes.Equal(i.stop.([]byte), b.BH.PrevHash) {
			break
		}
	}
	return result, nil
}
