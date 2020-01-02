package launcher

import (
	"database/sql"
	"git.ddex.io/infrastructure/ethereum-launcher/config"
	"git.ddex.io/infrastructure/ethereum-launcher/models"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestPickLaunchLogsPendingTooLongWithNoUrgent(t *testing.T) {
	_ = os.Setenv("ETHEREUM_NODE_URL", "http://localhost:8545")
	_ = os.Setenv("DATABASE_URL", "postgres://localhost:5432/launcher")
	_ = os.Setenv("MAX_GAS_PRICE_FOR_RETRY", "50")
	_ = os.Setenv("RETRY_PENDING_SECONDS_THRESHOLD", "90")
	_ = os.Setenv("RETRY_PENDING_SECONDS_THRESHOLD_FOR_URGENT", "60")

	config.InitConfig()

	var logs []*models.LaunchLog
	for i := 0; i <= 10; i++ {
		// -100, -85, -70, ..., 50
		log := &models.LaunchLog{
			Model: gorm.Model{ID: uint(i), CreatedAt: time.Now().Add(-100 * time.Second).Add(time.Duration(i*15) * time.Second)},
			Nonce: sql.NullInt64{Valid: true, Int64: int64(i)},
		}

		logs = append(logs, log)
	}

	resendingLogs := pickLaunchLogsPendingTooLong(logs)

	assert.Len(t, resendingLogs, 1)
	assert.Equal(t, uint(0), resendingLogs[0].ID)
}

func TestPickLaunchLogsPendingTooLongWithUrgent(t *testing.T) {
	_ = os.Setenv("ETHEREUM_NODE_URL", "http://localhost:8545")
	_ = os.Setenv("DATABASE_URL", "postgres://localhost:5432/launcher")
	_ = os.Setenv("MAX_GAS_PRICE_FOR_RETRY", "50")
	_ = os.Setenv("RETRY_PENDING_SECONDS_THRESHOLD", "90")
	_ = os.Setenv("RETRY_PENDING_SECONDS_THRESHOLD_FOR_URGENT", "60")

	config.InitConfig()

	var logs []*models.LaunchLog
	for i := 0; i <= 10; i++ {
		// -100, -85, -70, -55, ..., 50
		log := &models.LaunchLog{
			Model: gorm.Model{ID: uint(i), CreatedAt: time.Now().Add(-100 * time.Second).Add(time.Duration(i*15) * time.Second)},
			Nonce: sql.NullInt64{Valid: true, Int64: int64(i)},
		}

		logs = append(logs, log)
	}

	logs[2].IsUrgent = true

	resendingLogs := pickLaunchLogsPendingTooLong(logs)

	assert.Len(t, resendingLogs, 3)
	assert.Equal(t, uint(0), resendingLogs[0].ID)
	assert.Equal(t, uint(1), resendingLogs[1].ID)
	assert.Equal(t, uint(2), resendingLogs[2].ID)
}

func TestPickLaunchLogsPendingTooLongWhenNoLogs(t *testing.T) {
	_ = os.Setenv("ETHEREUM_NODE_URL", "http://localhost:8545")
	_ = os.Setenv("DATABASE_URL", "postgres://localhost:5432/launcher")
	_ = os.Setenv("MAX_GAS_PRICE_FOR_RETRY", "50")
	_ = os.Setenv("RETRY_PENDING_SECONDS_THRESHOLD", "90")
	_ = os.Setenv("RETRY_PENDING_SECONDS_THRESHOLD_FOR_URGENT", "60")

	config.InitConfig()

	var logs []*models.LaunchLog
	resendingLogs := pickLaunchLogsPendingTooLong(logs)

	assert.Len(t, resendingLogs, 0)
}

func TestNilSlice(t *testing.T) {
	nilSlice := returnNilSlice()
	assert.Nil(t, nilSlice)
	assert.Equal(t, 0, len(nilSlice))
}

func returnNilSlice() []int {
	return nil
}