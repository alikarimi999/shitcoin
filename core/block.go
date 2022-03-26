package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

func NewMsgdBlock(b *Block, sender NodeID, miner Address) *MsgBlock {
	mb := &MsgBlock{
		Sender: sender,
		Block:  b,
		Miner:  miner,
	}
	return mb
}

// This function Broadcast a new mined block in network
func (c *Chain) BroadBlock(mb *MsgBlock, cl http.Client) {

	for _, node := range c.KnownNodes {
		// dont send to miner of block or sender
		if mb.Sender == node.NodeId || NodeID(mb.Miner) == node.NodeId {
			continue
		}

		prev_sender := mb.Sender

		// Replace Message Block Sender ID with this node NodeID
		mb.Sender = NodeID(c.MinerAdd)
		b, _ := json.Marshal(mb)
		fmt.Printf("Sending block %x which received from Node %s to Node %s\n", mb.Block.BH.BlockHash, prev_sender, node.NodeId)
		cl.Post(fmt.Sprintf("%s/minedblock", node.FullAdd), "application/json", bytes.NewReader(b))
	}

}

func (c *Chain) AddBlockInDB(b *Block) {

	// Saving valid block in database
	err := saveBlockInDB(*b, &c.DB)
	if err != nil {
		log.Fatalf("... Block %x did not add to database\n\n", b.BH.BlockHash)
	}
	fmt.Printf("... Block %d with hash %x successfully added to database\n\n", b.BH.BlockIndex, b.BH.BlockHash)

}

// this function check validation of block that mined by another node
func (c *Chain) BlockValidator(b Block, ch *ChainState) bool {
	if b.BH.BlockIndex-1 == c.LastBlock.BH.BlockIndex && bytes.Equal(b.BH.PrevHash, c.LastBlock.BH.BlockHash) && b.Validate_hash() {

		if utxos, valid := c.Validate_transactions(b, ch); valid {
			ch.Utxos = utxos
			return true
		}

		return false
	}
	return false
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

// validate block's transactions
// and if transaction is valid update in memory UTXO set
func (c *Chain) Validate_transactions(b Block, ch *ChainState) (map[Account][]*UTXO, bool) {

	tempChainstate := &ChainState{}
	tempChainstate.Utxos = make(map[Account][]*UTXO)
	tempChainstate.Utxos = ch.Utxos

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
