package types

type UTXO struct {
	Txid  []byte
	Index uint
	Txout *TxOut
}
