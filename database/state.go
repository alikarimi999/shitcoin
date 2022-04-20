package database

import (
	"errors"
	"log"
	"time"

	"github.com/alikarimi999/shitcoin/core/types"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

func (d *database) SaveState(ss map[types.Account][]*types.UTXO, height uint64, wo *opt.WriteOptions) error {

	batch := new(leveldb.Batch)
	defer batch.Reset()
	for a, s := range ss {
		ts := tmplState{
			Chain_height: height,
			Time:         time.Now(),
			Owner:        []byte(a),
			Utxos:        s,
		}

		batch.Put(ts.Owner, Serialize(&ts))
	}
	err := d.db.Write(batch, wo)
	log.Println("save chain state on database")
	return err
}

func (d *database) ReadState() (map[types.Account][]*types.UTXO, error) {
	ss := make(map[types.Account][]*types.UTXO)

	iter := d.db.NewIterator(nil, nil)
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()

		d, err := Deserialize(value, &tmplState{})
		if err != nil {
			return nil, err
		}
		if ts, ok := d.(*tmplState); ok {
			ss[types.Account(key)] = ts.Utxos
		} else {
			return nil, errors.New("database:error on reading chain state from database")
		}

	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		return nil, err
	}

	return ss, nil

}
