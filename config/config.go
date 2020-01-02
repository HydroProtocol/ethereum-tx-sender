package config

import (
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"os"
	"strconv"
)

var Config *config

type config struct {
	DatabaseURL                           string          `json:"database_url"`
	MaxGasPriceForRetry                   decimal.Decimal `json:"max_gas_price_for_retry"`
	RetryPendingSecondsThreshold          int             `json:"retry_pending_seconds_threshold"`
	RetryPendingSecondsThresholdForUrgent int             `json:"retry_pending_seconds_threshold_for_urgent"`
	EthereumNodeUrl                       string          `json:"ethereum_node_url"`
	PrivateKeys                           string          `json:"PRIVATE_KEYS"`
}

func InitConfig() (*config, error) {
	databaseUrl := os.Getenv("DATABASE_URL")
	if databaseUrl == "" {
		return nil, errors.New("need DATABASE_URL env")
	}

	ethereumNodeUrl := os.Getenv("ETHEREUM_NODE_URL")
	if ethereumNodeUrl == "" {
		return nil, errors.New("need ETHEREUM_NODE_URL env")
	}

	maxGasPriceForRetry, err := decimal.NewFromString(os.Getenv("MAX_GAS_PRICE_FOR_RETRY"))
	if err != nil {
		return nil, fmt.Errorf("init RETRY_PENDING_SECONDS_THRESHOLD error, err %v", err)
	}

	retryPendingSecondsThreshold, err := strconv.ParseInt(os.Getenv("RETRY_PENDING_SECONDS_THRESHOLD"), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("init RETRY_PENDING_SECONDS_THRESHOLD error, err %v", err)
	}

	retryPendingSecondsThresholdForUrgent, err := strconv.ParseInt(os.Getenv("RETRY_PENDING_SECONDS_THRESHOLD_FOR_URGENT"), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("init RETRY_PENDING_SECONDS_THRESHOLD error, err %v", err)
	}

	Config = &config{
		EthereumNodeUrl:                       ethereumNodeUrl,
		DatabaseURL:                           databaseUrl,
		MaxGasPriceForRetry:                   maxGasPriceForRetry,
		RetryPendingSecondsThreshold:          int(retryPendingSecondsThreshold),
		RetryPendingSecondsThresholdForUrgent: int(retryPendingSecondsThresholdForUrgent),
	}

	return Config, nil
}
