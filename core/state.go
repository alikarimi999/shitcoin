package core

import (
	"bytes"
	"fmt"
	"log"
	"path/filepath"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/alikarimi999/shitcoin/core/types"
	"github.com/alikarimi999/shitcoin/database"
)

type chainstate interface {
	Handler()
	// if block was mined by this node local must true
	// if block mined by another node you must send a snapshot of block for preventing data race
	StateTransition(o any, local bool)
	GetTokens(account types.Account) []*types.UTXO
	GetMemTokens(account types.Account) []*types.UTXO
	GetStableSet() *types.UtxoSet

	// if miner start mining a block it will send a true to chainstate
	MinerIsStarting(start bool)
}

type stateTransmitter struct {
	TXs   []*types.Transaction
	local bool
}

type tmpUtxo struct {
	utxo      types.UTXO
	spendable bool
}

type MemSet struct {
	Mu     *sync.Mutex
	Tokens map[types.Account][]*tmpUtxo
}

func NewMemSet() *MemSet {
	return &MemSet{
		Mu:     &sync.Mutex{},
		Tokens: make(map[types.Account][]*tmpUtxo),
	}
}

type State struct {
	mu *sync.Mutex
	wg *sync.WaitGroup

	c *Chain

	memSet         *MemSet // temperory chain state
	memSetSnapshot *MemSet
	stableSet      *types.UtxoSet // chain state after bloack mined or remote mined block validate

	transportBlkCh  chan *stateTransmitter
	transportTxCh   chan *stateTransmitter
	minerstartingCh chan struct{}
	DB              database.Database
}

func NewState(c *Chain) *State {
	s := &State{
		mu:              &sync.Mutex{},
		wg:              c.Wg,
		c:               c,
		memSet:          NewMemSet(),
		memSetSnapshot:  NewMemSet(),
		stableSet:       types.NewUtxoSet(),
		transportBlkCh:  make(chan *stateTransmitter),
		transportTxCh:   make(chan *stateTransmitter),
		minerstartingCh: make(chan struct{}),
	}
	s.DB.SetupDB(filepath.Join(c.DBPath, "/chainstate"))

	return s
}

func (s *State) Handler() {
	s.wg.Add(1)
	defer s.wg.Done()
	log.Println("Chain State Handler start!!!")
	for {
		select {
		case t := <-s.transportBlkCh:
			s.mu.Lock()
			if !t.local {
				for _, tx := range t.TXs {
					s.stableSet.UpdateUtxoSet(tx)
				}
			} else {

				for a, utxos := range s.memSetSnapshot.Tokens {
					var acc int
					for _, utxo := range utxos {
						acc += utxo.utxo.Txout.Value
					}

					fmt.Printf("Len 2 >>>> %s >> %d  acc is %d\n", a, len(utxos), acc)
				}

				s.stableSet = s.memSetSnapshot.ConvertMem2Stable()
				s.memSetSnapshot.Tokens = make(map[types.Account][]*tmpUtxo)

			}
			s.saveInDB()
			s.memSet.SyncMemSet(s.stableSet.Tokens)
			s.mu.Unlock()
		case t := <-s.transportTxCh:
			s.mu.Lock()
			for _, tx := range t.TXs {
				s.memSet.UpdateMemSet(tx)
			}
			s.mu.Unlock()
		case <-s.minerstartingCh:
			s.mu.Lock()
			s.memSetSnapshot = s.memSet.SnapShot()
			for a, utxos := range s.memSetSnapshot.Tokens {
				var acc int
				for _, utxo := range utxos {
					acc += utxo.utxo.Txout.Value
				}

				fmt.Printf("Len 1 >>>> %s >> %d  acc is %d\n", a, len(utxos), acc)
			}
			if atomic.LoadUint64(&s.c.ChainHeight) > 0 { // if is not for genesis blockCh
				s.c.TxPool.ContinueHandler(true)
			}
			s.mu.Unlock()

		}
	}
}

func (s *State) GetTokens(account types.Account) []*types.UTXO {

	s.mu.Lock()
	s.memSet.Mu.Lock()
	defer s.memSet.Mu.Unlock()
	defer s.mu.Unlock()
	utxos := []*types.UTXO{}

	for _, tu := range s.memSet.Tokens[account] {
		if tu.spendable {
			utxos = append(utxos, &tu.utxo)
		}
	}

	return utxos
}

func (s *State) GetMemTokens(account types.Account) []*types.UTXO {

	s.mu.Lock()
	s.memSet.Mu.Lock()
	defer s.memSet.Mu.Unlock()
	defer s.mu.Unlock()
	utxos := []*types.UTXO{}

	for _, tu := range s.memSet.Tokens[account] {
		if tu.spendable {
			utxos = append(utxos, &tu.utxo)
		}
	}
	return utxos
}

func (s *State) GetStableSet() *types.UtxoSet {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stableSet.SnapShot()
}

func (s *State) StateTransition(o any, local bool) {

	switch t := o.(type) {
	case *types.Block:
		st := &stateTransmitter{local: local}
		if !local {
			for _, tx := range t.Transactions {
				st.TXs = append(st.TXs, tx)
			}
		}
		s.transportBlkCh <- st
		return
	case *types.Transaction:
		st := &stateTransmitter{}
		st.TXs = append(st.TXs, t)
		s.transportTxCh <- st
		return
	default:
		return
	}
}

func (s *State) saveInDB() {

	for account, utxos := range s.stableSet.Tokens {
		key := []byte(account)
		value := Serialize(utxos)

		err := s.DB.DB.Put(key, value, nil)

		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("All Tokens for %s saved in database\n\n", account)

	}

}

func (s *State) MinerIsStarting(start bool) {
	if start {
		s.minerstartingCh <- struct{}{}
	}
}

func (ms *MemSet) UpdateMemSet(tx *types.Transaction) {

	ms.Mu.Lock()
	defer ms.Mu.Unlock()
	var sender types.Account

	if !tx.IsCoinbase() {
		pk := tx.TxInputs[0].PublicKey
		account := types.Account(types.Pub2Address(pk, false))
		sender = account
		for _, in := range tx.TxInputs {

			for i, tu := range ms.Tokens[account] {
				if bytes.Equal(in.OutPoint, tu.utxo.Txid) && in.Vout == tu.utxo.Index && in.Value == tu.utxo.Txout.Value {
					ms.Tokens[account] = append(ms.Tokens[account][:i], ms.Tokens[account][i+1:]...)
					fmt.Printf("One Token with %d Value deleted from %s in memory\n ", tu.utxo.Txout.Value, types.Pub2Address(tu.utxo.Txout.PublicKeyHash, true))
					continue
				}
			}

		}

	}

	// add new Token
	var pkh []byte
	for index, out := range tx.TxOutputs {
		if out.Value == 0 {
			continue
		}
		pkh = out.PublicKeyHash
		account := types.Account(types.Pub2Address(pkh, true))

		tu := &tmpUtxo{
			utxo: types.UTXO{
				Txid:  tx.TxID,
				Index: uint(index),
				Txout: out,
			},
		}
		if sender == account {
			// tu.local = true
			tu.spendable = true
		}

		ms.Tokens[account] = append(ms.Tokens[account], tu)
		fmt.Printf("One Token with %d value added for %s in memory and spendable is(%v)\n", tu.utxo.Txout.Value, types.Pub2Address(tu.utxo.Txout.PublicKeyHash, true), tu.spendable)
	}

}

func (ms *MemSet) SyncMemSet(tokens map[types.Account][]*types.UTXO) {
	for a, tus := range ms.Tokens {
		for _, utxo := range tokens[a] {
			for _, tu := range tus {
				if reflect.DeepEqual(*utxo, tu.utxo) {
					tu.spendable = true
				}
			}
		}
	}

}

func (ms *MemSet) ConvertMem2Stable() *types.UtxoSet {

	us := &types.UtxoSet{
		Mu:     &sync.Mutex{},
		Tokens: make(map[types.Account][]*types.UTXO),
	}
	for a, tu := range ms.Tokens {
		for _, u := range tu {
			us.Tokens[a] = append(us.Tokens[a], &u.utxo)
		}
	}
	return us
}

func (ms *MemSet) SnapShot() *MemSet {
	m := &MemSet{
		Mu:     &sync.Mutex{},
		Tokens: make(map[types.Account][]*tmpUtxo),
	}

	for a, utxos := range ms.Tokens {
		for _, tu := range utxos {
			tmp := &tmpUtxo{
				utxo:      *tu.utxo.SnapShot(),
				spendable: tu.spendable,
			}
			m.Tokens[a] = append(m.Tokens[a], tmp)
		}
	}

	return m
}

func ConvertStable2Mem(us *types.UtxoSet) *MemSet {
	ms := &MemSet{
		Mu:     &sync.Mutex{},
		Tokens: make(map[types.Account][]*tmpUtxo),
	}

	for a, utxos := range us.Tokens {

		for _, utxo := range utxos {
			tmp := &tmpUtxo{
				utxo:      *utxo,
				spendable: true,
			}
			ms.Tokens[a] = append(ms.Tokens[a], tmp)
		}
	}

	return ms
}
