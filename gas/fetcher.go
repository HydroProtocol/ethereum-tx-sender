package gas

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/HydroProtocol/hydro-sdk-backend/utils"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

var prices *GasPrices
var fetchPriceMutex sync.Mutex

func GetCurrentGasPrice() (decimal.Decimal,decimal.Decimal) {
	if prices == nil {
		fetch(context.Background())
		if prices == nil {
			panic("Can't get gas price, will use default")
		}
	}

	return prices.Proposed, prices.Proposed.Mul(decimal.NewFromFloat(1.1))
}

func StartFetcher(ctx context.Context) {
	logrus.Infof("gas fetcher started")
	fetch(ctx)
	for {
		select {
		case <-ctx.Done():
			logrus.Infof("gas fetcher exit")
			return
		case <-time.After(10 * time.Second):
			fetch(ctx)
		}
	}
}

func fetch(ctx context.Context) {
	fetchPriceMutex.Lock()
	defer fetchPriceMutex.Unlock()
	price, err := fetchPrice(ctx)
	if err != nil {
		logrus.Error(err)
	} else {
		logrus.Infof("fetch price: %s", utils.ToJsonString(price))
	}
}

func fetchPrice(ctx context.Context) (*EtherGasStationResponse, error){
	r, _ := http.NewRequest(http.MethodGet, "https://ethgasstation.info/json/ethgasAPI.json", nil)
	r = r.WithContext(ctx)

	resp, err := http.DefaultClient.Do(r)

	if err != nil {
		return nil, fmt.Errorf("fetch price from ether gas station failed err: %v", err)
	}

	resBytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("read bytes err: %v", err)
	}

	var egsRes EtherGasStationResponse
	err = json.Unmarshal(resBytes, &egsRes)
	if err != nil {
		return nil, err
	} else {
		prices = getPrices(&egsRes)
	}

	return &egsRes, nil
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

var Gwei = decimal.New(1, 9)                // Gwei
var EtherGasStationUnit = decimal.New(1, 8) // 0.1 Gwei
var maxGasPrice = decimal.New(100, 9)       // 100 Gwei
var minGasPrice = decimal.New(1, 9)         // 1 Gwei
