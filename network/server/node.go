package server

import (
	"fmt"

	"github.com/alikarimi999/shitcoin/core/types"
	netype "github.com/alikarimi999/shitcoin/network/types"
	"github.com/labstack/echo/v4"
)

func (s *Server) SendNodeInfo(ctx echo.Context) error {
	node := s.Ch.Node
	err := ctx.Bind(node)
	if err != nil {
		return err
	}
	ctx.JSONPretty(200, node, " ")
	return nil
}

func (s *Server) SendNodes(ctx echo.Context) error {
	gn := &netype.GetNode{}

	err := ctx.Bind(gn)
	if err != nil {
		return err
	}

	sender := NETWORKPROTO + ctx.RealIP() + gn.SrcNodes[0].Port
	senderID := gn.SrcNodes[0].ID
	gn.SrcNodes[0].FullAdd = sender
	fmt.Printf("Node %s with Address %s Requesitng new node\n", gn.SrcNodes[0].ID, gn.SrcNodes[0].FullAdd)

	s.Ch.NMU.Lock()
	gn.ShareNodes = s.collectNodes(gn.SrcNodes, senderID)

	for _, n := range gn.SrcNodes {
		if len(s.Ch.Peers) >= 8 {
			break
		}
		if _, ok := s.Ch.Peers[n.ID]; !ok && n.ID != s.Ch.Node.ID {
			s.Ch.Peers[n.ID] = n
			fmt.Printf("...Add Node %s with address %s to Peers\n", n.ID, n.FullAdd)
		}

		// FIXME:
		// h := atomic.LoadUint64(&s.Ch.ChainHeight)
		// if h < n.NodeHeight {
		// 	fmt.Printf("... Node %s had %d mined block more\n", n.ID, n.NodeHeight-h)
		// 	fmt.Printf("... Trying to sync with Node %s\n", n.ID)
		// 	s.Ch.Mu.Lock()

		// 	// s.Client.Sync(n)
		// 	s.Ch.Mu.Unlock()
		// }
	}
	defer s.Ch.NMU.Unlock()
	ctx.JSONPretty(200, gn, " ")
	return nil
}

func (s *Server) collectNodes(src []*types.Node, sender string) []*types.Node {
	share_nodes := []*types.Node{}

	// first node that any node share to other nodes refers to itself
	n := s.Ch.Node
	share_nodes = append(share_nodes, n)
	fmt.Printf("...Sending Node %s\n", n.ID)

Out:
	for _, node := range s.Ch.Peers {

		// every node share 1 Peers an itself address
		// this made applicant node to requests other nodes for sharing their nodes too
		// and this made the network more destributed
		if len(share_nodes) >= 2 {
			break
		}

		// dont share node if applicant node already has it
		for _, n := range src {
			if node.ID == n.ID {
				fmt.Printf("... Node %s already has this Node %s with address %s so don't send it\n", sender, n.ID, n.FullAdd)
				continue Out
			}
		}
		if node.ID == sender {
			continue Out
		}
		share_nodes = append(share_nodes, node)
		fmt.Printf("...Sending Node %s with Address %s\n", node.ID, node.FullAdd)

	}
	return share_nodes
}
