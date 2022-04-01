package core

type memPool struct {
	Transactions []*Transaction
	Chainstate   *ChainState
}

// Transfer Transactions from transaction pool to Block
func (mp *memPool) TransferTxs2Block(b *Block, miner Address, amount int) error {
	mp.addMinerReward(miner, amount)

	b.Transactions = mp.Transactions

	// and make transaction pool Clean
	mp.Clean()

	return nil

}

func (mp *memPool) addMinerReward(miner Address, amount int) error {
	pkh := Add2PKH(miner)
	reward := CoinbaseTx(pkh, amount)
	mp.Transactions = append(mp.Transactions, reward)
	mp.Chainstate.UpdateUtxoSet(reward)

	return nil
}

// Clean transaction pool
func (mp *memPool) Clean() error {
	mp.Transactions = []*Transaction{}

	return nil
}

func (mp *memPool) SnapShot() *memPool {

	pool := &memPool{
		Chainstate: mp.Chainstate.SnapShot(),
	}

	for _, tx := range mp.Transactions {
		pool.Transactions = append(pool.Transactions, tx.SnapShot())
	}
	return pool
}
