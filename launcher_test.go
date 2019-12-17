package main

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestPickLaunchLogsPendingTooLongWithNoUrgent(t *testing.T) {
	config = &Config{}
	config.RetryPendingSecondsThreshold = 90
	config.RetryPendingSecondsThresholdForUrgent = 60

	var logs []*LaunchLog
	for i := 0; i <= 10; i++ {
		// -100, -85, -70, ..., 50
		log := &LaunchLog{
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
	config = &Config{}
	config.RetryPendingSecondsThreshold = 90
	config.RetryPendingSecondsThresholdForUrgent = 60

	var logs []*LaunchLog
	for i := 0; i <= 10; i++ {
		// -100, -85, -70, -55, ..., 50
		log := &LaunchLog{
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
	config = &Config{}
	config.RetryPendingSecondsThreshold = 90
	config.RetryPendingSecondsThresholdForUrgent = 60

	var logs []*LaunchLog
	resendingLogs := pickLaunchLogsPendingTooLong(logs)

	assert.Len(t, resendingLogs, 0)
}

func TestEmptySlice(t *testing.T) {
	var emptySlice []int
	assert.Nil(t, emptySlice)
	assert.Equal(t, 0, len(emptySlice))

	for i := range emptySlice {
		logrus.Infoln(i)
	}

	emptySlice = returnEmptySlice()
	assert.NotNil(t, emptySlice)
}

func returnEmptySlice() []int {
	return []int{}
}
