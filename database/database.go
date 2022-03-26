package database

import (
	"log"

	leveldb "github.com/syndtr/goleveldb/leveldb"
)

type Database struct {
	DB *leveldb.DB
}

func (d *Database) SetupDB(path string) {

	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		log.Fatalln(err)
	}

	d.DB = db

}
