package core

import (
	"fmt"
	"log"
	"sync"

	"github.com/alikarimi999/shitcoin/core/types"
	"github.com/alikarimi999/shitcoin/database"
)

type chainstate interface {
	Handler()
	// if block was mined by this node local must true
	// if block mined by another node you must send a snapshot of block for preventing data race
	StateTransition(o any, local bool)
	GetTokens(account types.Account) []*types.UTXO
	GetStableSet() *types.UtxoSet

	// if miner start mining a block it will send a true to chainstate
	MineStarted(start bool)
}

type stateTransmitter struct {
	TXs       []*types.Transaction
	fromBlock bool // it shows transactions come from a mined block or just recieved from network
	local     bool
}

type State struct {
	wg *sync.WaitGroup

	memSet     *types.UtxoSet // temperory chain state
	stableSet  *types.UtxoSet // chain state after bloack mined or remote mined block validate
	pendingSet *types.UtxoSet // when mining process of a block start

	transportBlkCh chan *stateTransmitter
	transportTxCh  chan *stateTransmitter
	startmineCh    chan struct{}
	DB             database.Database
}

func NewState(dbPath string, wg *sync.WaitGroup) *State {
	s := &State{
		wg:             wg,
		memSet:         types.NewUtxoSet(),
		stableSet:      types.NewUtxoSet(),
		pendingSet:     types.NewUtxoSet(),
		transportBlkCh: make(chan *stateTransmitter),
		transportTxCh:  make(chan *stateTransmitter),
		startmineCh:    make(chan struct{}),
	}
	s.DB.SetupDB(dbPath)

	return s
}

func (s *State) Handler() {
	s.wg.Add(1)
	defer s.wg.Done()
	log.Println("Chain State Handler start!!!")
	for {
		select {
		case t := <-s.transportBlkCh:
			if t.fromBlock {
				if !t.local {
					for _, tx := range t.TXs {
						s.stableSet.UpdateUtxoSet(tx)
					}
					s.memSet = s.stableSet
				} else {
					s.stableSet = s.pendingSet
					s.pendingSet.Clean()
				}
				go s.saveInDB()
			}

		case t := <-s.transportTxCh:
			for _, tx := range t.TXs {
				s.memSet.UpdateUtxoSet(tx)
			}
		case <-s.startmineCh:
			s.pendingSet = s.memSet
			s.memSet.Clean()

		}
	}
}

func (s *State) GetTokens(account types.Account) []*types.UTXO {
	return s.memSet.Tokens[account]
}

func (s *State) GetStableSet() *types.UtxoSet {
	return s.stableSet.SnapShot()
}

func (s *State) StateTransition(o any, local bool) {

	switch t := o.(type) {
	case *types.Block:
		st := &stateTransmitter{fromBlock: true, local: local}
		if !local {
			for _, tx := range t.Transactions {
				st.TXs = append(st.TXs, tx)
			}
		}
		s.transportBlkCh <- st
		return
	case *types.Transaction:
		st := &stateTransmitter{fromBlock: false}
		st.TXs = append(st.TXs, t)
		s.transportTxCh <- st
		return
	default:
		return
	}
}

func (s *State) saveInDB() {
	s.stableSet.Mu.Lock()
	defer s.stableSet.Mu.Unlock()
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

func (s *State) MineStarted(start bool) {
	if start {
		s.startmineCh <- struct{}{}
	}
}
