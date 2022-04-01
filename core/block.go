package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
)

type blockHeader struct {
	Timestamp  int64
	PrevHash   []byte
	BlockIndex uint64
	BlockHash  []byte
	Miner      Address
	Nonce      uint64
	Difficulty uint64
}

type Block struct {
	BH           *blockHeader
	Transactions []*Transaction
}

func NewBlock() *Block {
	return &Block{
		BH: &blockHeader{
			Timestamp:  0,
			PrevHash:   []byte{},
			BlockIndex: 0,
			BlockHash:  []byte{},
			Miner:      []byte{},
			Nonce:      0,
			Difficulty: 0,
		},
		Transactions: make([]*Transaction, 0),
	}
}

func (b *Block) Serialize() []byte {

	d := bytes.Join(
		[][]byte{
			b.BH.PrevHash,
			Int2Hex(int64(b.BH.BlockIndex)),
			Int2Hex(int64(b.BH.Difficulty)),
			b.BH.Miner,
			Int2Hex(int64(b.BH.Nonce)),
			Int2Hex(b.BH.Timestamp),
			SerializeTxs(b.Transactions),
		},
		nil,
	)

	return d
}

func (b *Block) Hash() []byte {
	b.BH.BlockHash = nil
	data := b.Serialize()

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

// =================================================================================================================
// =================================================================================================================

func (c *Chain) AddBlockInDB(b *Block) {

	// Saving valid block in database
	err := b.SaveBlockInDB(&c.DB)
	if err != nil {
		log.Fatalf("... Block %x did not add to database\n\n", b.BH.BlockHash)
	}
	fmt.Printf("... Block %d with hash %x successfully added to database\n\n", b.BH.BlockIndex, b.BH.BlockHash)

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
