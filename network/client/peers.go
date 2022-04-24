package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/alikarimi999/shitcoin/core/types"
	netype "github.com/alikarimi999/shitcoin/network/types"
)

const (
	MaxNodesQueue = netype.MaxPeers
)

type nodesQueue struct {
	nodes chan *types.Node
}

func newNodesQueue(size int) *nodesQueue {
	return &nodesQueue{
		nodes: make(chan *types.Node, size),
	}
}

func (n *nodesQueue) push(node *types.Node) {
	select {
	case n.nodes <- node:
	default:
		// todo: right this shit
		log.Println("nodes queue is full")
	}
}

func (n *nodesQueue) pop() *types.Node {
	select {
	case node := <-n.nodes:
		return node
	default:
		log.Println("nodes queue is empty ")
		return &types.Node{}
	}
}

func (c *Client) Peers(dst string) error {

	node, err := nodeInfo(dst, c.Cl)
	if err != nil {
		log.Fatalln(err)
	}
	c.PeerSet.Add(node)

	err = c.sharePeers(dst)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) getNewNodes(dst string) []*types.Node {

	src_nodes := []*types.Node{}

	// first element in slice always refer to node itself
	src_nodes = append(src_nodes, c.Ch.Node)

	for _, n := range c.PeerSet.Peers {
		src_nodes = append(src_nodes, n)
	}
	gn := &netype.GetNode{
		SrcNodes:   src_nodes,
		ShareNodes: nil,
	}

	b, _ := json.Marshal(gn)
	resp, err := http.Post(fmt.Sprintf("%s/getnode", dst), "application/json", bytes.NewReader(b))
	if err != nil {
		fmt.Println(err.Error())
		return src_nodes
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		return src_nodes
	}
	gn = new(netype.GetNode)
	err = json.Unmarshal(body, gn)
	if err != nil {
		fmt.Println(err.Error())
		return src_nodes
	}

	// first DstNodes always is the node that share it's Peers with this node
	gn.ShareNodes[0].FullAdd = dst
	return gn.ShareNodes
}

// This function help to obtain some node addresses of network and add these PeerSet
func (c *Client) sharePeers(dst string) error {

	// nodes that we didn't ask to share their nodes with us yet
	nq := newNodesQueue(MaxNodesQueue)

Out:
	for i := 0; i <= netype.MaxPeers; i++ {
		log.Printf("requesting new peers from %s\n", dst)

		if len(c.PeerSet.Peers) >= netype.MaxPeers {
			return fmt.Errorf("peerset is full")
		}
		share_nodes := c.getNewNodes(dst)
		if len(share_nodes) == 0 {
			log.Printf("node %s didn't share any peer", dst)
			continue
		}

		for _, n := range share_nodes {

			fmt.Printf("peer %s recieved from %s\n", n.ID, dst)

			if len(c.PeerSet.Peers) >= netype.MaxPeers {
				break Out
			}

			// dont add if n refers to this node
			if n.ID == c.Ch.Node.ID {
				continue
			}
			if _, ok := c.PeerSet.Peers[n.ID]; ok {
				fmt.Println("peer already exist in peer set")
				continue
			}

			c.PeerSet.Add(n)
			fmt.Printf("peer %s with address %s added to peerset\n", n.ID, n.FullAdd)

			if n.FullAdd != dst {
				nq.push(n)
			}
		}

		dst = nq.pop().FullAdd
		if dst == "" {
			break
		}

	}

	return nil
}

func nodeInfo(dst string, cl http.Client) (*types.Node, error) {

	node := &types.Node{}
	resp, err := cl.Get(fmt.Sprintf("%s/nodeinfo", dst))
	if err != nil {
		return node, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return node, err
	}
	err = json.Unmarshal(body, node)
	if err != nil {
		return node, err
	}
	node.FullAdd = dst
	return node, nil

}
