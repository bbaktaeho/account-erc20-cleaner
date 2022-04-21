package main

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"log"
	"math/big"
	"os"
	"vault-to-ledger/erc20"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	rpcUrl          string
	vaultPrivateKey *ecdsa.PrivateKey
	ledgerAddress   string
	erc20Address    string
)

func init() {
	log.Println("init start")
	defer log.Println("init end")

	rpcUrl = os.Getenv("RPC_URL")
	vaultPrivateHexKey := os.Getenv("VAULT_PRIVATE_KEY")
	ledgerAddress = os.Getenv("LEDGER_ADDRESS")
	erc20Address = os.Getenv("ERC20_ADDRESS")

	c, err := ethclient.Dial(rpcUrl)
	if err != nil {
		log.Println("rpc url error")
		log.Fatal(err)
	}
	defer c.Close()

	v, err := crypto.HexToECDSA(vaultPrivateHexKey)
	if err != nil {
		log.Println("vault private key error")
		log.Fatal(err)
	}
	vaultPrivateKey = v

	if !common.IsHexAddress(ledgerAddress) {
		log.Fatal(errors.New("ledger address error"))
	}

	if !common.IsHexAddress(erc20Address) {
		log.Fatal(errors.New("ledger address error"))
	}
}

func main() {
	ctx := context.Background()
	client, _ := ethclient.Dial(rpcUrl)

	publickey, ok := vaultPrivateKey.Public().(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publickey)
	toAddress := common.HexToAddress(ledgerAddress)
	tokenAddress := common.HexToAddress(erc20Address)

	kok, err := erc20.NewIERC20(tokenAddress, client)
	if err != nil {
		log.Fatal(err)
	}

	balance, err := kok.BalanceOf(nil, fromAddress)
	if err != nil {
		log.Fatal(balance)
	}
	log.Println("ERC-20 balance of vault:", balance.String())

	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		log.Fatal(err)
	}

	signer := bind.NewKeyedTransactor(vaultPrivateKey)
	signer.Nonce = big.NewInt(int64(nonce))
	signer.Value = big.NewInt(0)
	signer.GasLimit = uint64(100000)
	signer.GasPrice = gasPrice.Add(gasPrice, gasPrice.Div(gasPrice, big.NewInt(10)))

	tx, err := kok.Transfer(signer, toAddress, balance)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("transaction hash:", tx.Hash().Hex())
}
