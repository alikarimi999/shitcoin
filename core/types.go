package core

import "github.com/alikarimi999/shitcoin/core/types"

type any interface{}

type txid string

type Transactions map[txid]*types.Transaction
