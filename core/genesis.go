package core

import (
	"github.com/alikarimi999/shitcoin/core/types"
)

// Creat genesis Block
func (c *Chain) creatGenesis() {

	pkh := types.Add2PKH(c.MinerAdd)
	tx := types.CoinbaseTx(pkh, 15)

	c.Miner.MineGenesis(tx)

}
