package database

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/alikarimi999/shitcoin/core/types"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

func (d *database) SaveBlock(block *types.Block, sender, miner string, wo *opt.WriteOptions) error {

	tb := tmplBlk{
		Miner:  miner,
		Sender: sender,
		Time:   time.Now(),
		Block:  *block,
	}
	b := Serialize(tb)
	err := d.db.Put(block.BH.BlockHash, b, wo)
	if err != nil {
		return err
	}

	err = d.db.Put([]byte(fmt.Sprintf("block_index_%d", block.BH.BlockIndex)), block.BH.BlockHash, wo)
	if err != nil {
		return err
	}
	err = d.db.Put([]byte("last_hash"), block.BH.BlockHash, wo)
	if err != nil {
		return err
	}
	log.Printf("block %x saved in database\n", block.BH.BlockHash)
	return nil
}

func (d *database) LastBlock(ro *opt.ReadOptions) (*types.Block, error) {
	hash, err := d.db.Get([]byte("last_hash"), ro)
	if err != nil || hash == nil {
		return nil, errors.New("database: database is empty")
	}
	return d.GetBlockH(hash, ro)

}

// retrieve block by hash
func (d *database) GetBlockH(hash []byte, ro *opt.ReadOptions) (*types.Block, error) {
	b, err := d.db.Get(hash, ro)
	if err != nil {
		return nil, err
	}

	bd, err := Deserialize(b, &tmplBlk{})
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	if tb, ok := bd.(*tmplBlk); ok {
		return &tb.Block, nil

	}
	return nil, errors.New("couldn't access to block")

}

// retrieve block by Index
func (d *database) GetBlockI(index uint64, ro *opt.ReadOptions) (*types.Block, error) {
	hash, err := d.GetBlkHash(index, ro)
	if err != nil {
		return nil, err
	}

	return d.GetBlockH(hash, ro)
}

func (d *database) GetBlkHash(index uint64, ro *opt.ReadOptions) ([]byte, error) {
	return d.db.Get([]byte(fmt.Sprintf("block_index_%d", index)), ro)
}
