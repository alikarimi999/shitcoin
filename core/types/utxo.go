package types

type UTXO struct {
	Txid  []byte
	Index uint
	Txout *TxOut
}

func (u *UTXO) SnapShot() *UTXO {

	out := *u.Txout
	ut := &UTXO{
		Txid:  u.Txid,
		Index: u.Index,
		Txout: &out,
	}
	return ut
}
