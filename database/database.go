package database

import (
	"log"

	"github.com/alikarimi999/shitcoin/core/types"
	leveldb "github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type DB interface {
	SaveBlock(block *types.Block, sender, miner string, wo *opt.WriteOptions) error
	LastBlock(ro *opt.ReadOptions) (*types.Block, error)
	// retrieve block by hash
	GetBlockH(hash []byte, ro *opt.ReadOptions) (*types.Block, error)
	// retrieve block by index
	GetBlockI(index uint64, ro *opt.ReadOptions) (*types.Block, error)
	// retrieve block hash by index
	GetBlkHash(index uint64, ro *opt.ReadOptions) ([]byte, error)

	SaveState(ss map[types.Account][]*types.UTXO, height uint64, wo *opt.WriteOptions) error
}
type database struct {
	db *leveldb.DB
}

func SetupDB(path string) *database {

	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		log.Fatalln(err)
	}

	return &database{db: db}

}
