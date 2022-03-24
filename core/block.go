package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"strings"
)

type blockHeader struct {
	Timestamp  int64
	PrevHash   []byte
	BlockIndex uint64
	BlockHash  []byte
	Nonce      uint64
	Difficulty uint64
}

type Block struct {
	BH           *blockHeader
	Transactions []*Transaction
}

func NewBlock() *Block {
	return &Block{
		BH:           &blockHeader{},
		Transactions: make([]*Transaction, 0),
	}
}

func (b *Block) Print() {
	fmt.Printf("\n%s\n", strings.Repeat("=", 100))

	fmt.Printf("\nBlock: %d\nHash: %x\n\n", b.BH.BlockIndex, b.BH.BlockHash)
	for _, tx := range b.Transactions {
		tx.Print()
	}
	fmt.Printf("\n%s\n", strings.Repeat("=", 100))

}

func (b *Block) Serialize() []byte {

	d := bytes.Join(
		[][]byte{
			b.BH.PrevHash,
			Int2Hex(int64(b.BH.BlockIndex)),
			Int2Hex(int64(b.BH.Difficulty)),
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

func (c *Chain) AddNewBlock(b *Block) bool {

	if c.BlockValidator(*b) {
		fmt.Printf("Block %x is valid\n", b.BH.BlockHash)
		err := SaveBlockInDB(*b, &c.DB)
		if err != nil {
			log.Fatalf("Block %x did not add to database\n\n", b.BH.BlockHash)
		}
		fmt.Printf("Block %d with hash %x successfully added to database\n\n", b.BH.BlockIndex, b.BH.BlockHash)

		c.SyncUtxoSet()
		return true
	}
	fmt.Printf("Block %x is not valid\n", b.BH.BlockHash)
	return false
}

// this function check validation of block that mined by another node
func (c *Chain) BlockValidator(b Block) bool {

	if b.BH.BlockIndex-1 == c.LastBlock.BH.BlockIndex && b.Validate_hash() {

		if utxos, valid := c.Validate_transactions(b); valid {
			c.MemPool.Chainstate.Utxos = utxos
		} else {
			return false
		}

		return true
	}
	return false
}

// if BlockValidator was true then node sync it's chain state base on block transactions inputs and outputs
func (c *Chain) update_utxo_set(b Block) {
	fmt.Println("Updateing UTXO set")
	for _, tx := range b.Transactions {
		c.MemPool.Chainstate.UpdateUtxoSet(tx)
	}
}

func (b *Block) Validate_hash() bool {

	hash := b.BH.BlockHash
	real_hash := b.Hash()
	b.BH.BlockHash = hash

	result := bytes.Equal(hash, real_hash)
	if result {
		fmt.Println(" Block hash is valid")
	} else {
		fmt.Println(" Block hash is not valid")

	}
	return result

}

// validate block's transactions
// and if transaction is valid update in memory UTXO set
func (c *Chain) Validate_transactions(b Block) (map[Account][]*UTXO, bool) {

	tempChainstate := &ChainState{}
	tempChainstate.Utxos = make(map[Account][]*UTXO)
	tempChainstate.Utxos = c.MemPool.Chainstate.Utxos

	for _, tx := range b.Transactions {

		if tx.IsCoinbase() {
			tempChainstate.UpdateUtxoSet(tx)

			continue
		}
		trx := *tx
		if !tempChainstate.Verifyhash(tx) || !trx.Checksig() {
			return nil, false
		}
		tempChainstate.UpdateUtxoSet(tx)
	}

	return tempChainstate.Utxos, true
}
