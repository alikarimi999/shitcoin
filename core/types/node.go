package types

import "fmt"

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

// return a Node struct that represent this node
func ThisNode(nid []byte, port int, last_hash []byte, height uint64) *Node {

	n := &Node{
		NodeId:     NodeID(nid),
		FullAdd:    "",
		Port:       fmt.Sprintf(":%d", port),
		LastHash:   last_hash,
		NodeHeight: height,
	}

	return n
}
