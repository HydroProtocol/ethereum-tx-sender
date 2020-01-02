package pkm

import (
	"crypto/ecdsa"
	"fmt"
	"git.ddex.io/infrastructure/ethereum-launcher/config"
	"git.ddex.io/lib/ethrpc"
	"github.com/HydroProtocol/hydro-sdk-backend/sdk/crypto"
	"github.com/sirupsen/logrus"
	"strings"
)

type Pkm interface {
	Sign(t *ethrpc.T) (string, error)
}

type localPKM struct {
	KeyMap map[string]*ecdsa.PrivateKey
}

var LocalPKM *localPKM

func init() {
	privateKeys := config.Config.PrivateKeys
	privateKeyList := strings.Split(privateKeys, ",")

	keyMap := make(map[string]*ecdsa.PrivateKey)
	for idx, privateKeyHex := range privateKeyList {
		privateKey, err := crypto.NewPrivateKeyByHex(privateKeyHex)
		if err != nil {
			logrus.Errorf("parse private key fail, key is num %d", idx)
		}
		publicKey := crypto.PubKey2Address(privateKey.PublicKey)
		keyMap[publicKey] = privateKey
		logrus.Infof("parse private key success, public key is %s", publicKey)
	}
	LocalPKM = &localPKM{
		KeyMap: keyMap,
	}
}

func (l localPKM) Sign(t *ethrpc.T) (string, error) {
	privateKey, ok := l.KeyMap[t.From]
	if !ok {
		return "", fmt.Errorf("cannot sign by account %s", t.From)
	}

	//privateKey

	return "", nil
}

//
//type PKMResponse struct {
//	Status       bool            `json:"success"`
//	Data         PKMResponseData `json:"data"`
//	ErrorMessage string          `json:"errMsg"`
//}
//
//type PKMResponseData struct {
//	TransactionId string `json:"transaction_id"`
//	RawData       string `json:"raw_data"`
//}
//
//func pkmSign(t *ethrpc.T) (string, error) {
//	bts, _ := json.Marshal(map[string]interface{}{
//		"from":      t.From,
//		"to":        t.To,
//		"data":      t.Data,
//		"gas_price": t.GasPrice,
//		"amount":    t.Value,
//		"gas_limit": t.Gas,
//		"nonce":     t.Nonce,
//	})
//
//	res, err := http.Post(config.PkmUrl+"/signTransaction", "application/json", bytes.NewReader(bts))
//
//	if err != nil {
//		return "", err
//	}
//
//	retStr, _ := ioutil.ReadAll(res.Body)
//
//	if len(retStr) == 0 {
//		return "", fmt.Errorf("empty sign result")
//	}
//
//	var resp PKMResponse
//
//	_ = json.Unmarshal([]byte(retStr), &resp)
//
//	if !resp.Status {
//		return "", fmt.Errorf("sign result error %s", resp.ErrorMessage)
//	}
//
//	return resp.Data.RawData, nil
//}
