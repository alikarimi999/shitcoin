package main

import (
	"os"

	"github.com/alikarimi999/shitcoin/cli"
)

func main() {

	defer os.Exit(0)
	cmd := new(cli.Commandline)

	cmd.Run()
}

// func send(c *blockchain.Chain, w *wallet.Wallet, to string, amount float64) error {
// 	from := w.Address()
// 	tx, err := blockchain.NewTx(c, from, []byte(to), amount)
// 	if err != nil {
// 		return err
// 	}
// 	tx, err = w.SignTX(tx)
// 	Handle(err)
// 	err = c.AddTx2Pool(tx)
// 	Handle(err)
// 	fmt.Println("Transaction was successfull!")
// 	return nil
// }
