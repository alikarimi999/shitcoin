package types

type Chainid int

type publickey []byte

type Node struct {
	NodeId NodeID
	// Node full address
	FullAdd    string
	Port       string
	LastHash   []byte
	NodeHeight uint64
}

// every node hash a NodeID which is miner Address of that node
type NodeID Account

// account address in byte
type Address []byte

// Account address in string
type Account string
