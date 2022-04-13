package types

import (
	"crypto/rand"
	"fmt"
	"log"

	"github.com/alikarimi999/shitcoin/config"
)

type Node struct {
	ID string
	// Node full address
	FullAdd     string
	Port        string
	GenesisHash []byte
	LastHash    []byte
	NodeHeight  uint64
}

func NodeID(config config.Config) string {

	id := config.GetNodeID()
	if id == "" {
		var i [32]byte
		rand.Read(i[:])
		id = fmt.Sprintf("%x", i[:])
		config.SetNodeID(id)
		err := config.SaveConfig()
		if err != nil {
			log.Fatal(err)
		}
	}
	return id
}

func NewNode(config config.Config, port int, last_hash []byte, height uint64) *Node {
	return &Node{
		ID:          NodeID(config),
		FullAdd:     "",
		Port:        fmt.Sprintf(":%d", port),
		GenesisHash: []byte{},
		LastHash:    last_hash,
		NodeHeight:  height,
	}
}
