package config

import (
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestInitConfig(t *testing.T) {
	_ = os.Setenv("ETHEREUM_NODE_URL", "http://localhost:8545")
	_ = os.Setenv("DATABASE_URL", "postgres://localhost:5432/launcher")
	_ = os.Setenv("MAX_GAS_PRICE_FOR_RETRY", "50")
	_ = os.Setenv("RETRY_PENDING_SECONDS_THRESHOLD", "10")
	_ = os.Setenv("RETRY_PENDING_SECONDS_THRESHOLD_FOR_URGENT", "5")
	_ = os.Setenv("PRIVATE_KEYS", "private key")

	launcherConfig,err := InitConfig()
	assert.Nil(t, err)
	assert.EqualValues(t, "http://localhost:8545", launcherConfig.EthereumNodeUrl)
	assert.EqualValues(t, "postgres://localhost:5432/launcher", launcherConfig.DatabaseURL)
	assert.EqualValues(t, decimal.New(50, 0), launcherConfig.MaxGasPriceForRetry)
	assert.EqualValues(t, 10, launcherConfig.RetryPendingSecondsThreshold)
	assert.EqualValues(t, 5, launcherConfig.RetryPendingSecondsThresholdForUrgent)

	assert.EqualValues(t, "http://localhost:8545", Config.EthereumNodeUrl)
	assert.EqualValues(t, "postgres://localhost:5432/launcher", Config.DatabaseURL)
	assert.EqualValues(t, decimal.New(50, 0), Config.MaxGasPriceForRetry)
	assert.EqualValues(t, 10, Config.RetryPendingSecondsThreshold)
	assert.EqualValues(t, 5, Config.RetryPendingSecondsThresholdForUrgent)
}
