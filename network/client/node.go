package client

import (
	"fmt"

	"github.com/alikarimi999/shitcoin/core/types"
)

const (
	MaxPeers int = 8
	MaxNodesQueue
)

type NodesQueue struct {
	nodes chan *types.Node
}

func NewNodesQueue(size int) *NodesQueue {
	return &NodesQueue{
		nodes: make(chan *types.Node, size),
	}
}

func (n *NodesQueue) Push(node *types.Node) {
	select {
	case n.nodes <- node:
	default:
		// todo: right this shit
		fmt.Println("Nodes Queue is full")
	}
}

func (n *NodesQueue) Pop() *types.Node {
	select {
	case node := <-n.nodes:
		return node
	default:
		fmt.Println("Nodes Queue is empty ")
		return &types.Node{}
	}
}
