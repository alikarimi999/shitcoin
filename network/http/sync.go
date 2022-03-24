package network

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/alikarimi999/shitcoin/core"
)

func Sync(c *Objects, cl http.Client) {

	vs := GetVersion(c.Ch.KnownNodes, cl)
	// sync node is node with best chain
	syncNode := &Version{}

	for j := 0; j < len(vs); j++ {
		if syncNode.NodeHeight < vs[j].NodeHeight {
			syncNode = vs[j]
		}
	}

	if syncNode.NodeHeight == c.Ch.ChainHeight {
		if !bytes.Equal(syncNode.LastHash, c.Ch.LastBlock.BH.BlockHash) {
			fmt.Printf(" Node \"%s\" and your node are not in same network please connect to another node\n", syncNode.Address)
			c.Ch.DeleteNode(syncNode.Address)
			return
		}
		fmt.Println("Node is Updated!")
	}

	if syncNode.NodeHeight > c.Ch.ChainHeight {

		fmt.Printf("Sync node is %s with %d chain height\n", syncNode.Address, syncNode.NodeHeight)

		if c.Ch.ChainHeight == 0 {
			block := getGen(syncNode.Address, cl)

			c.Ch.MemPool.Chainstate.UpdateUtxoSet(block.Transactions[0])
			err := core.SaveGenInDB(*block, &c.Ch.DB)
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Printf("Genesis Block added to database\n")
			c.Ch.LastBlock = block
			c.Ch.ChainHeight++
		}
		bh, err := getData(c, syncNode.Address, cl)
		if err != nil {
			fmt.Println(err.Error())
		}

		for i := 0; i < len(bh); i++ {
			hash := bh[blockIndex(c.Ch.LastBlock.BH.BlockIndex+1)]

			block := getBlock(hash, syncNode.Address, cl)
			if block == nil {
				break
			}
			fmt.Printf("Block %x Downloaded\n", block.BH.BlockHash)
			if !c.Ch.AddNewBlock(block) {
				break
			}
			c.Ch.LastBlock = block

		}
	}
}

func getBlock(hash []byte, syncAddress string, cl http.Client) *core.Block {
	data := GetBlock{hash}
	msg, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	resp, err := cl.Post(fmt.Sprintf("%s/getblock", syncAddress), "application/json", bytes.NewReader(msg))
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	b := new(core.Block)
	err = json.Unmarshal(body, b)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return b

}

// get genesis block
func getGen(syncNode string, cl http.Client) *core.Block {
	resp, err := cl.Get(fmt.Sprintf("%s/getgen", syncNode))
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	block := new(core.Block)
	json.Unmarshal(body, block)
	if block.Validate_hash() {
		fmt.Printf("Genesis Block downloaded\n")

		return block

	}
	fmt.Println("Genesis block is not valid")
	return nil

}

func getData(c *Objects, syncAddress string, cl http.Client) (map[blockIndex][]byte, error) {

	data := GetData{c.Ch.LastBlock.BH.BlockHash}
	b, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err.Error())
		return nil, nil
	}
	resp, err := cl.Post(fmt.Sprintf("%s/getdata", syncAddress), "application/json", bytes.NewReader(b))
	if err != nil {
		fmt.Println(err.Error())
		return nil, nil

	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		return nil, nil

	}
	inv := new(Inv)
	json.Unmarshal(body, inv)
	if inv.InvType == blockType {
		return inv.BlocksHash, nil
	}
	return nil, errors.New("sync node response currupt data to getData function")

}

func GetVersion(nodes []string, cl http.Client) []*Version {
	var vs []*Version

	for _, node := range nodes {
		resp, err := cl.Get(fmt.Sprintf("%s/getver", node))
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		v := new(Version)
		json.Unmarshal(body, v)
		v.Address = node
		vs = append(vs, v)
	}

	return vs
}
