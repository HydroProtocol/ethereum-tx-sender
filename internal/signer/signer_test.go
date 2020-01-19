package signer

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/onrik/ethrpc"
	"math/big"
	"testing"
)

func TestLocalPKM_Sign(t *testing.T) {
	InitPKM("60210631062ca6cdf438253a4bb57dedd5ebb8079fa2234c65644a4c9600c52b")

	transaction := ethrpc.T{
		From:     "0xcd7be3e30eedc0699281b4eb35e7d8afa560b773",
		To:       "0xcd7be3e30eedc0699281b4eb35e7d8afa560b773",
		Gas:      100000,
		GasPrice: big.NewInt(9000000000),
		Value:    big.NewInt(10000000000000000),
		Data:     "0x",
		Nonce:    5,
	}

	raw,_:= LocalSigner.Sign(&transaction)
	spew.Dump(raw)


	ethrpcClient := ethrpc.New("https://ropsten.infura.io/v3/19d753b2600445e292d54b1ef58d4df4")
	spew.Dump(ethrpcClient.EthSendRawTransaction(raw))
}
