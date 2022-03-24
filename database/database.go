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

// func CheckDB(path string) bool {

// 	if file, err := os.Stat(path); err != nil {
// 		if os.IsNotExist(err) {
// 			log.Fatalln(errors.New(fmt.Sprintf("Database does not exist in %s", path)))
// 		}
// 		log.Fatalln(err)

// 	} else {
// 		if !file.IsDir() {
// 			log.Fatalln(errors.New(fmt.Sprintf("Database does not exist in %s", path)))
// 		}

// 	}

// 	ls, err := os.ReadDir(path)
// 	if err != nil {
// 		log.Fatalln(err)
// 	}

// 	for _, f := range ls {
// 		fn := f.Name()
// 		if len(fn) >= 8 && fn[:8] == "MANIFEST" {
// 			return true
// 		}
// 	}

// 	return false
// }
