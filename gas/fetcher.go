package gas

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

var prices *GasPrices

type EtherGasStationResponse struct {
	Fast        decimal.Decimal `json:"fast"`
	Fastest     decimal.Decimal `json:"fastest"`
	SafeLow     decimal.Decimal `json:"safeLow"`
	Average     decimal.Decimal `json:"average"`
	BlockTime   decimal.Decimal `json:"block_time"`
	BlockNum    decimal.Decimal `json:"blockNum"`
	Speed       decimal.Decimal `json:"speed"`
	SafeLowWait decimal.Decimal `json:"safeLowWait"`
	AvgWait     decimal.Decimal `json:"avgWait"`
	FastWait    decimal.Decimal `json:"fastWait"`
	FastestWait decimal.Decimal `json:"fastestWait"`
}

type GasPrices struct {
	Proposed decimal.Decimal `json:"proposed"`
	Low      decimal.Decimal `json:"low"`
	Average  decimal.Decimal `json:"average"`
	High     decimal.Decimal `json:"high"`
}

func Get() (*GasPrices, error) {
	if prices == nil {
		return nil,errors.New("bad gas")
	}

	return prices,nil
}

func GetCurrentGasPrice(isUrgent bool) decimal.Decimal {
	currentGas, err := Get()

	if err != nil {
		logrus.Error("Can't get gas price, will use default")
	}

	if isUrgent {
		// 1.1 times high
		return currentGas.Proposed.Mul(decimal.NewFromFloat(1.1))
	}

	return currentGas.Proposed
}

var Gwei = decimal.New(1, 9)                // Gwei
var EtherGasStationUnit = decimal.New(1, 8) // 0.1 Gwei
var maxGasPrice = decimal.New(100, 9)       // 100 Gwei
var minGasPrice = decimal.New(1, 9)         // 1 Gwei

func StartFetcher(ctx context.Context) {
	firstTime := true

	for {
		if !firstTime {
			select {
			case <-ctx.Done():
				return
			case <-time.After(10 * time.Second):
			}
		}

		firstTime = false

		r, _ := http.NewRequest(http.MethodGet, "https://ethgasstation.info/json/ethgasAPI.json", nil)
		r = r.WithContext(ctx)

		resp, err := http.DefaultClient.Do(r)

		if err != nil {
			logrus.Errorf("fetch price from ether gas station failed err: %v", err)
			continue
		}

		resBytes, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			logrus.Errorf("read bytes err: %v", err)
			continue
		}

		var egsRes EtherGasStationResponse
		_ = json.Unmarshal(resBytes, &egsRes)
		prices = getPrices(&egsRes)

		logrus.Infof("new price fetched: %v", string(resBytes))
	}
}

func getPrices(ethGasStationResp *EtherGasStationResponse) *GasPrices {
	high := ethGasStationResp.Fast.Mul(EtherGasStationUnit)
	average := ethGasStationResp.Average.Mul(EtherGasStationUnit)
	low := ethGasStationResp.SafeLow.Mul(EtherGasStationUnit)

	proposed := high.Add(Gwei.Mul(decimal.NewFromFloat(5)))

	return &GasPrices{
		Proposed: keepInSafeRange(proposed),
		High:     keepInSafeRange(high),
		Average:  keepInSafeRange(average),
		Low:      keepInSafeRange(low),
	}
}

func keepInSafeRange(d decimal.Decimal) decimal.Decimal {
	if d.GreaterThan(maxGasPrice) {
		return maxGasPrice
	}

	if d.LessThan(minGasPrice) {
		return minGasPrice
	}

	return d
}
