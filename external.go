package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"git.ddex.io/infrastructure/ethereum-gas-price/client"
	"git.ddex.io/lib/ethrpc"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

func getCurrentGasPrice() decimal.Decimal {
	prices, err := client.Get()

	if err != nil {
		logrus.Error("Can't get gas price, will use default")
	}

	return prices.Proposed
}

type PKMResponse struct {
	Status       bool            `json:"success"`
	Data         PKMResponseData `json:"data"`
	ErrorMessage string          `json:"error_message"`
}

type PKMResponseData struct {
	TransactionId string `json:"transaction_id"`
	RawData       string `json:"raw_data"`
}

func pkmSign(t *ethrpc.T) (string, error) {
	bts, _ := json.Marshal(map[string]interface{}{
		"from":      t.From,
		"to":        t.To,
		"data":      t.Data,
		"gas_price": t.GasPrice,
		"gas_limit": t.Gas,
		"nonce":     t.Nonce,
	})

	res, err := http.Post(config.PkmUrl+"/signTransaction", "application/json", bytes.NewReader(bts))

	if err != nil {
		return "", err
	}

	retStr, _ := ioutil.ReadAll(res.Body)

	if len(retStr) == 0 {
		return "", fmt.Errorf("empty sign result")
	}

	var resp PKMResponse

	_ = json.Unmarshal([]byte(retStr), &resp)

	if !resp.Status {
		return "", fmt.Errorf("sign result error %s", resp.ErrorMessage)
	}

	return resp.Data.RawData, nil
}
