package hdwallet

import (
	"crypto/ecdsa"
	"fmt"

	hd "github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip39"
)

type Wallet struct {
	masterKey *hd.ExtendedKey
}

type Derived struct {
	ExtendedKey *hd.ExtendedKey
	Address     common.Address
	PrivateKey  *ecdsa.PrivateKey
	PublicKey   ecdsa.PublicKey
}

// DefaultDerivationPath is the default derivation path used
var DefaultDerivationPath = "m/44'/60'/0'/0/%d"

// New creates a new Wallet instance from a mnemonic and optional password
func New(mnemonic, password string) (*Wallet, error) {
	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, password)
	if err != nil {
		return nil, err
	}

	masterKey, err := hd.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, err
	}

	return &Wallet{
		masterKey: masterKey,
	}, nil
}

// NewFromSeed creates a new Wallet instance from a seed
func NewFromSeed(seed []byte) (*Wallet, error) {
	masterKey, err := hd.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, err
	}

	return &Wallet{masterKey: masterKey}, nil
}

// CreateDerivationPath creates a derivation path from an index
func CreateDerivationPath(index int) (accounts.DerivationPath, error) {
	path := fmt.Sprintf(DefaultDerivationPath, index)
	parsed, err := accounts.ParseDerivationPath(path)
	if err != nil {
		return nil, err
	}

	return parsed, nil
}

// Derive derives the master key using the provided path.
func (w *Wallet) Derive(path accounts.DerivationPath) (*Derived, error) {
	key := w.masterKey
	var err error

	for _, n := range path {
		if key.IsAffectedByIssue172() {
			key, err = key.Derive(n)
		} else {
			key, err = key.DeriveNonStandard(n)
		}

		if err != nil {
			return nil, err
		}
	}

	privateKey, err := getPrivateKey(key)
	if err != nil {
		return nil, err
	}

	publicKey := privateKey.PublicKey

	return &Derived{
		ExtendedKey: key,
		Address:     crypto.PubkeyToAddress(publicKey),
		PrivateKey:  privateKey,
		PublicKey:   publicKey,
	}, nil
}

// Account returns an ethereum account.
func (d *Derived) Account() accounts.Account {
	return accounts.Account{Address: d.Address}
}

func getPrivateKey(key *hd.ExtendedKey) (*ecdsa.PrivateKey, error) {
	privKey, err := key.ECPrivKey()
	if err != nil {
		return nil, err
	}

	return privKey.ToECDSA(), nil
}
