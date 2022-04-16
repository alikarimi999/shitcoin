package database

import (
	"time"

	"github.com/alikarimi999/shitcoin/core/types"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

func (d *database) SaveState(ss map[types.Account][]*types.UTXO, height uint64, wo *opt.WriteOptions) error {

	for a, s := range ss {
		ts := tmplState{
			Chain_height: height,
			Time:         time.Now(),
			Owner:        []byte(a),
			Utxos:        s,
		}
		err := d.db.Put(ts.Owner, Serialize(&ts), wo)
		if err != nil {
			return err
		}
	}

	return nil
}
