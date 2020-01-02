package pkm

import (
	"fmt"
	"git.ddex.io/lib/ethrpc"
)

type Pkm interface {
	Sign(t *ethrpc.T) (string, error)
}

type localPKM struct{
	KeyMap map[string] string
}

var LocalPKM *localPKM

func init(){
	LocalPKM = &localPKM{

	}
}
func (l localPKM) Sign(t *ethrpc.T) (string, error) {
	if _, ok := l.KeyMap[t.From]; !ok {
		return "", fmt.Errorf("cannot sign by account %s", t.From)
	}

	return "",nil
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
