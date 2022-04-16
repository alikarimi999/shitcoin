package core

// type DatabaseIterator struct {
// 	NextHash []byte
// 	DB       database.Database
// }

// func (c *Chain) NewIter() *DatabaseIterator {

// 	return &DatabaseIterator{c.LastBlock.BH.BlockHash, c.DB}

// }

// func (iter *DatabaseIterator) Next() (*types.Block, error) {
// 	block, err := ReadBlock(iter.DB, iter.NextHash)
// 	if err != nil {
// 		return nil, err
// 	}
// 	iter.NextHash = block.BH.PrevHash
// 	return block, nil
// }

// func ReadBlock(d database.Database, hash []byte) (*types.Block, error) {

// 	b, err := d.DB.Get(hash, nil)
// 	if err != nil {
// 		return nil, errors.New("can't get block from database")
// 	}
// 	bl := Deserialize(b, new(types.Block))

// 	if block, ok := bl.(*types.Block); ok {
// 		return block, nil

// 	}

// 	return nil, errors.New("can't get block from database")
// }

// func Serialize(t interface{}) []byte {
// 	buff := bytes.Buffer{}

// 	encoder := gob.NewEncoder(&buff)
// 	encoder.Encode(t)

// 	return buff.Bytes()
// }

// func Deserialize(b []byte, t interface{}) interface{} {

// 	decoder := gob.NewDecoder(bytes.NewBuffer(b))

// 	err := decoder.Decode(t)

// 	if err != nil {
// 		log.Fatalln(err)
// 	}

// 	return t
// }
