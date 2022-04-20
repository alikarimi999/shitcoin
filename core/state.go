package core

import (
	"bytes"
	"log"
	"path/filepath"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alikarimi999/shitcoin/core/types"
	"github.com/alikarimi999/shitcoin/database"
)

type chainstate interface {
	Handler(wg *sync.WaitGroup)
	// if block was mined by this node local must true
	// if block mined by another node you must send a snapshot of block for preventing data race
	StateTransition(o any, local bool)
	GetTokens(account types.Account) []*types.UTXO
	GetMemTokens(account types.Account) []*types.UTXO
	GetStableSet() *types.UtxoSet

	// if miner start mining a block it will send a true to chainstate
	MinerIsStarting(start bool)

	GenesisUpdate(b *types.Block)

	// Load chain state from database
	Load() error
}

type stateTransmitter struct {
	TXs   []*types.Transaction
	local bool
}

type tmpUtxo struct {
	time      time.Duration
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

	c *Chain

	memSet         *MemSet // temperory chain state
	memSetSnapshot *MemSet
	stableSet      *types.UtxoSet // chain state after bloack mined or remote mined block validate

	transportBlkCh chan *stateTransmitter
	transportTxCh  chan *stateTransmitter

	minerstartingCh chan struct{}
	DB              database.DB
}

func NewState(c *Chain) *State {
	s := &State{
		mu:              &sync.Mutex{},
		c:               c,
		memSet:          NewMemSet(),
		memSetSnapshot:  NewMemSet(),
		stableSet:       types.NewUtxoSet(),
		transportBlkCh:  make(chan *stateTransmitter),
		transportTxCh:   make(chan *stateTransmitter),
		minerstartingCh: make(chan struct{}),
	}
	s.DB = database.SetupDB(filepath.Join(c.DBPath, "/chainstate"))

	return s
}

func (s *State) Handler(wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("Chain State Handler start!!!")
	for {
		select {
		case t := <-s.transportBlkCh:
			s.mu.Lock()
			if !t.local {
				for _, tx := range t.TXs {
					s.stableSet.UpdateUtxoSet(tx)
				}
				s.memSet = convert(s.stableSet)

				st := <-s.transportTxCh // recieve remaining transactions from transaction pool
				for _, tx := range st.TXs {
					s.memSet.update(tx)
				}

			} else {

				tx := t.TXs[0]
				owner := types.Account(s.c.MinerAdd)
				utxo := types.UTXO{
					Txid:  tx.TxID,
					Index: 0,
					Txout: tx.TxOutputs[0],
				}

				s.memSetSnapshot.Tokens[owner] = append(s.memSetSnapshot.Tokens[owner], &tmpUtxo{utxo: utxo, spendable: true})
				s.memSet.Tokens[owner] = append(s.memSet.Tokens[owner], &tmpUtxo{utxo: utxo, spendable: true})
				s.stableSet = s.memSetSnapshot.ConvertMem2Stable()
			}
			s.save()

			s.SyncMemSet(s.stableSet.Tokens)
			s.memSetSnapshot.Tokens = make(map[types.Account][]*tmpUtxo)
			s.mu.Unlock()
		case t := <-s.transportTxCh:
			s.mu.Lock()
			for _, tx := range t.TXs {
				s.memSet.update(tx)
			}
			s.mu.Unlock()
		case <-s.minerstartingCh:
			s.mu.Lock()
			s.memSetSnapshot = s.memSet.SnapShot()
			s.c.TxPool.ContinueHandler(true)
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
		} else {
			// just send minerReward transation to state handler
			// miner reward transaction is last one
			for _, tx := range t.Transactions {
				if tx.IsCoinbase() {
					st.TXs = append(st.TXs, tx)
					break
				}
			}

		}

		s.transportBlkCh <- st
		return
	case *types.Transaction:
		st := &stateTransmitter{}
		st.TXs = append(st.TXs, t)
		s.transportTxCh <- st
		return
	case []*types.Transaction:
		st := &stateTransmitter{}

		for _, tx := range t {
			st.TXs = append(st.TXs, tx)
		}
		s.transportTxCh <- st
		return
	default:
		return
	}
}

func (s *State) save() {
	err := s.DB.SaveState(s.stableSet.Tokens, atomic.LoadUint64(&s.c.ChainHeight), nil)
	if err != nil {
		log.Fatalln(err)
	}
}

func (s *State) MinerIsStarting(start bool) {
	if start {
		s.minerstartingCh <- struct{}{}
	}
}

func (ms *MemSet) update(tx *types.Transaction) {

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
					// fmt.Printf("One Token with %d Value deleted from %s in memory\n ", tu.utxo.Txout.Value, types.Pub2Address(tu.utxo.Txout.PublicKeyHash, true))
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
			time: time.Duration(time.Now().UnixNano()),
			utxo: types.UTXO{
				Txid:  tx.TxID,
				Index: uint(index),
				Txout: out,
			},
		}
		if sender == account {
			tu.spendable = true
		}

		ms.Tokens[account] = append(ms.Tokens[account], tu)
		// fmt.Printf("One Token with %d value added for %s in memory and spendable is(%v)\n", tu.utxo.Txout.Value, types.Pub2Address(tu.utxo.Txout.PublicKeyHash, true), tu.spendable)
	}

}

func (s *State) SyncMemSet(tokens map[types.Account][]*types.UTXO) {
	for a, tus := range s.memSet.Tokens {
		for _, utxo := range tokens[a] {
			for _, tu := range tus {
				if tu.time < s.c.Miner.StartTime() && reflect.DeepEqual(*utxo, tu.utxo) {
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

func convert(stable *types.UtxoSet) *MemSet {
	ms := &MemSet{
		Mu:     &sync.Mutex{},
		Tokens: make(map[types.Account][]*tmpUtxo),
	}

	for a, utxos := range stable.Tokens {

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

func (s *State) GenesisUpdate(b *types.Block) {
	for _, tx := range b.Transactions {
		s.memSet.update(tx)
		s.memSet.Tokens[types.Account(s.c.MinerAdd)][0].spendable = true
		s.stableSet.UpdateUtxoSet(tx)
	}
	s.save()
}

func (s *State) Load() error {
	ss, err := s.DB.ReadState()
	if err != nil {
		return err
	}
	s.stableSet.Tokens = ss
	s.memSet = convert(s.stableSet)
	return nil
}
