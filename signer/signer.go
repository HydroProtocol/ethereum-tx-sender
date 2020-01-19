package signer

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/HydroProtocol/hydro-sdk-backend/sdk/crypto"
	"github.com/HydroProtocol/hydro-sdk-backend/sdk/signer"
	"github.com/HydroProtocol/hydro-sdk-backend/sdk/types"
	"github.com/HydroProtocol/hydro-sdk-backend/utils"
	"github.com/onrik/ethrpc"
	"github.com/sirupsen/logrus"
	"strings"
)

type Signer interface {
	Sign(t *ethrpc.T) (string, error)
}

type localSigner struct {
	KeyMap map[string]*ecdsa.PrivateKey
}

var LocalSigner *localSigner

func InitPKM(privateKeys string) Signer {
	privateKeyList := strings.Split(privateKeys, ",")

	keyMap := make(map[string]*ecdsa.PrivateKey)
	for idx, privateKeyHex := range privateKeyList {
		privateKey, err := crypto.NewPrivateKeyByHex(privateKeyHex)
		if err != nil {
			logrus.Errorf("parse private key fail, key is num %d", idx)
			continue
		}

		publicKey := crypto.PubKey2Address(privateKey.PublicKey)
		keyMap[publicKey] = privateKey
		logrus.Infof("parse private key success, public key is %s", publicKey)
	}

	LocalSigner = &localSigner{
		KeyMap: keyMap,
	}

	return LocalSigner
}

func (l localSigner) Sign(t *ethrpc.T) (string, error) {
	privateKey, ok := l.KeyMap[t.From]
	if !ok {
		return "", fmt.Errorf("cannot sign by account %s", t.From)
	}

	tx := types.NewTransaction(
		uint64(t.Nonce),
		t.To,
		t.Value,
		uint64(t.Gas),
		t.GasPrice,
		utils.Hex2Bytes(t.Data),
	)
	fmt.Println(utils.ToJsonString(tx))
	signedTransaction, err := signer.SignTx(tx, privateKey)

	if err != nil {
		utils.Errorf("sign transaction error: %v", err)
		panic(err)
	}

	return utils.Bytes2HexP(signer.EncodeRlp(signedTransaction)), nil
}
