package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"log"
	"sync"

	"github.com/alikarimi999/shitcoin/database"
)

type BlockHeader struct {
	Timestamp  int64
	PrevHash   []byte
	BlockIndex uint64
	BlockHash  []byte
	Miner      Address
	Nonce      uint64
	Difficulty uint64
}

type Block struct {
	BH           *BlockHeader
	Transactions []*Transaction
}

func NewBlock() *Block {
	h := [32]byte{}
	return &Block{
		BH: &BlockHeader{
			Timestamp:  0,
			PrevHash:   []byte{},
			BlockIndex: 0,
			BlockHash:  h[:],
			Miner:      []byte{},
			Nonce:      0,
			Difficulty: 0,
		},
		Transactions: make([]*Transaction, 0),
	}

}

func (b *Block) Validate_hash() bool {

	hash := b.BH.BlockHash
	real_hash := b.Hash()
	b.BH.BlockHash = hash

	result := bytes.Equal(hash, real_hash)
	if result {
		fmt.Println("... Block hash is valid")
	} else {
		fmt.Println("... Block hash is not valid")

	}
	return result

}

func (b Block) SaveBlockInDB(d *database.Database, mu *sync.Mutex) error {

	mu.Lock()
	block := Serialize(b)
	mu.Unlock()

	key := b.BH.BlockHash
	value := block

	err := d.DB.Put(key, value, nil)
	if err != nil {
		fmt.Println(err.Error())
	}

	err = d.DB.Put([]byte("last_hash"), b.BH.BlockHash, nil)
	fmt.Printf("Last Block in database is %x\n\n", b.BH.BlockHash)
	return err
}

// a deep copy of Block
func (b *Block) SnapShot() *Block {

	bh := *b.BH
	nb := &Block{
		BH: &bh,
	}

	for _, tx := range b.Transactions {
		t := *tx
		nb.Transactions = append(nb.Transactions, &t)
	}

	return nb
}

func (b *Block) serialize() []byte {
	d := bytes.Join(
		[][]byte{
			b.BH.PrevHash,
			Int2Hex(int64(b.BH.BlockIndex)),
			Int2Hex(int64(b.BH.Difficulty)),
			b.BH.Miner,
			Int2Hex(int64(b.BH.Nonce)),
			Int2Hex(b.BH.Timestamp),
			join(b.Transactions),
		},
		nil,
	)

	return d
}

func (b *Block) Hash() []byte {
	b.BH.BlockHash = nil
	data := b.serialize()

	hash := sha256.Sum256(data)
	return hash[:]
}

func Int2Hex(n int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, n)
	if err != nil {
		log.Fatalln(err)
	}

	return buff.Bytes()
}
func Serialize(t interface{}) []byte {
	buff := bytes.Buffer{}

	encoder := gob.NewEncoder(&buff)
	encoder.Encode(t)

	return buff.Bytes()
}

func Deserialize(b []byte, t interface{}) interface{} {

	decoder := gob.NewDecoder(bytes.NewBuffer(b))

	err := decoder.Decode(t)

	if err != nil {
		log.Fatalln(err)
	}

	return t
}
