package types

import (
	"log"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/base58"
)

var (
	params = chaincfg.Params{
		PubKeyHashAddrID: 0x00,
		HDPrivateKeyID:   [4]byte{0x04, 0x88, 0xad, 0xe4}, // starts with xprv
		HDPublicKeyID:    [4]byte{0x04, 0x88, 0xb2, 0x1e}, // starts with xpub
		// // BIP44 coin type used in the hierarchical deterministic path for
		// // address generation.
		HDCoinType: 0,
	}
)

func Add2PKH(address Address) []byte {
	pkh, _, err := base58.CheckDecode(string(address))
	if err != nil {
		log.Fatal(err)
	}
	return pkh
}

func PK2Add(pk []byte, isHash bool) string {
	if !isHash {
		add, err := btcutil.NewAddressPubKey(pk, &params)
		if err != nil {
			log.Fatal(err)
		}
		return add.EncodeAddress()
	}

	add, err := btcutil.NewAddressPubKeyHash(pk, &params)
	if err != nil {
		log.Fatal(err)
	}
	return add.EncodeAddress()

}

// func PK2Add(pk []byte) Account {
// 	publicKeyHash := Hash160(pk)
// 	versionedHahs := AddVersion(publicKeyHash, byte(0x00))
// 	check := Checksum(versionedHahs)

// 	cHash := append(versionedHahs, check...)

// 	address := base58.Encode(cHash)

// 	return Account(address)

// }

// func Checksum(b []byte) []byte {
// 	h1 := sha256.Sum256(b)
// 	h2 := sha256.Sum256(h1[:])

// 	return h2[:4]
// }

// this function add addressVersion
// func AddVersion(b []byte, v byte) []byte {

// 	return append([]byte{v}, b...)
// }

// this is obvoious
// func Hash160(pub []byte) []byte {
// 	hash := sha256.Sum256(pub)

// 	hasher := ripemd160.New()
// 	_, err := hasher.Write(hash[:])
// 	if err != nil {
// 		log.Panic(err)
// 	}

// 	pkh := hasher.Sum(nil)

// 	return pkh
// }

// Conver Pub Key to Coin address

// func Pub2Address(pub []byte, hash bool) []byte {

// 	if !hash {
// 		pub = Hash160(pub)
// 	}
// 	versionedHahs := AddVersion(pub, byte(0x00))
// 	check := Checksum(versionedHahs)

// 	cHash := append(versionedHahs, check...)

// 	address := base58.Encode(cHash)

// 	return []byte(address)
// }
