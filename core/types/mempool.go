package types

import "sync"

type MemPool struct {
	Mu           *sync.Mutex
	Transactions []*Transaction
	Chainstate   *ChainState
}

// Transfer Transactions from transaction pool to Block
func (mp *MemPool) TransferTxs2Block(b *Block, miner Address, amount int) error {
	mp.addMinerReward(miner, amount)

	for _, tx := range mp.Transactions {
		b.Transactions = append(b.Transactions, tx.SnapShot())
	}
	return nil

}

func (mp *MemPool) addMinerReward(miner Address, amount int) error {
	pkh := Add2PKH(miner)
	reward := CoinbaseTx(pkh, amount)
	mp.Transactions = append(mp.Transactions, reward)
	mp.Chainstate.UpdateUtxoSet(reward)

	return nil
}

// Clean transaction pool
func (mp *MemPool) Clean() error {
	mp.Transactions = []*Transaction{}

	return nil
}

func (mp *MemPool) SnapShot() *MemPool {

	pool := &MemPool{
		Chainstate: mp.Chainstate.SnapShot(),
	}

	for _, tx := range mp.Transactions {
		pool.Transactions = append(pool.Transactions, tx.SnapShot())
	}
	return pool
}