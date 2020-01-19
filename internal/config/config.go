package config

import (
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
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
	PrivateKeys                           string          `json:"private_keys"`
}

func InitConfig() (*config, error) {
	databaseUrl := os.Getenv("DATABASE_URL")
	if databaseUrl == "" {
		return nil, errors.New("missing environment variable: DATABASE_URL")
	}

	ethereumNodeUrl := os.Getenv("ETHEREUM_NODE_URL")
	if ethereumNodeUrl == "" {
		return nil, errors.New("missing environment variable: ETHEREUM_NODE_URL")
	}

	privateKeys := os.Getenv("PRIVATE_KEYS")
	if privateKeys == "" {
		return nil, fmt.Errorf("missing environment variable: PRIVATE_KEYS")
	}

	maxGasPriceForRetryStringValue := os.Getenv("MAX_GAS_PRICE_FOR_RETRY")
	maxGasPriceForRetry, err := decimal.NewFromString(maxGasPriceForRetryStringValue)
	if maxGasPriceForRetryStringValue == "" || err != nil {
		maxGasPriceForRetry = decimal.New(50, 9)
		logrus.Infof("MAX_GAS_PRICE_FOR_RETRY use default value %s", maxGasPriceForRetry)
	}

	retryPendingSecondsThresholdStringValue := os.Getenv("RETRY_PENDING_SECONDS_THRESHOLD")
	retryPendingSecondsThreshold, err := strconv.ParseInt(retryPendingSecondsThresholdStringValue, 10, 32)
	if err != nil {
		retryPendingSecondsThreshold = 90
		logrus.Infof("RETRY_PENDING_SECONDS_THRESHOLD use default value %s", maxGasPriceForRetry)
	}

	retryPendingSecondsThresholdForUrgentStringValue := os.Getenv("RETRY_PENDING_SECONDS_THRESHOLD_FOR_URGENT")
	retryPendingSecondsThresholdForUrgent, err := strconv.ParseInt(retryPendingSecondsThresholdForUrgentStringValue, 10, 32)
	if err != nil {
		retryPendingSecondsThresholdForUrgent = 45
		logrus.Infof("RETRY_PENDING_SECONDS_THRESHOLD_FOR_URGENT use default value %s", maxGasPriceForRetry)
	}

	Config = &config{
		EthereumNodeUrl:                       ethereumNodeUrl,
		DatabaseURL:                           databaseUrl,
		MaxGasPriceForRetry:                   maxGasPriceForRetry,
		RetryPendingSecondsThreshold:          int(retryPendingSecondsThreshold),
		RetryPendingSecondsThresholdForUrgent: int(retryPendingSecondsThresholdForUrgent),
		PrivateKeys:                           privateKeys,
	}

	return Config, nil
}
