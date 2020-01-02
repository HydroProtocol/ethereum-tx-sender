package main

import (
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestInitConfig(t *testing.T) {
	_ = os.Setenv("DATABASE_URL", "postgres://localhost:5432/launcher")
	_ = os.Setenv("MAX_GAS_PRICE_FOR_RETRY", "50")
	_ = os.Setenv("RETRY_PENDING_SECONDS_THRESHOLD", "10")
	_ = os.Setenv("RETRY_PENDING_SECONDS_THRESHOLD_FOR_URGENT", "5")

	launcherConfig,err := InitConfig()
	assert.Nil(t, err)
	assert.EqualValues(t, "postgres://localhost:5432/launcher", launcherConfig.DatabaseURL)
	assert.EqualValues(t, decimal.New(50, 0), launcherConfig.MaxGasPriceForRetry)
	assert.EqualValues(t, 10, launcherConfig.RetryPendingSecondsThreshold)
	assert.EqualValues(t, 5, launcherConfig.RetryPendingSecondsThresholdForUrgent)
}
