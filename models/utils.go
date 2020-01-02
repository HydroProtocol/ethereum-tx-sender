package models

import (
	"git.ddex.io/infrastructure/ethereum-launcher/messages"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"math/rand"
	"time"
)

func HandleLaunchLogStatus(log *LaunchLog, result bool, gasUsed int, executedAt int) (*LaunchLog, error) {
	var statusCode messages.LaunchLogStatus

	if result {
		statusCode = messages.LaunchLogStatus_SUCCESS
	} else {
		statusCode = messages.LaunchLogStatus_FAILED
	}

	status := messages.LaunchLogStatus_name[int32(statusCode)]

	log.Status = status
	log.GasUsed = uint64(gasUsed)
	log.ExecutedAt = uint64(executedAt)

	err := ExecuteInRepeatableReadTransaction(func(tx *gorm.DB) (err error) {
		var reloadedLog LaunchLog

		if err = tx.Model(&reloadedLog).Set("gorm:query_option", "FOR UPDATE").Where("id = ?", log.ID).Scan(&reloadedLog).Error; err != nil {
			return err
		}

		if reloadedLog.Status != messages.LaunchLogStatus_PENDING.String() &&
				reloadedLog.Status != messages.LaunchLogStatus_CREATED.String() {
			return nil
		}

		if err = tx.Model(LaunchLog{}).Where(
			"item_type = ? and item_id = ? and status = ? and hash != ?",
			log.ItemType,
			log.ItemID,
			messages.LaunchLogStatus_PENDING.String(),
			log.Hash,
		).Update(map[string]interface{}{
			"status": messages.LaunchLogStatus_RETRIED.String(),
		}).Error; err != nil {
			logrus.Errorf("set retry status failed log: %+v err: %+v", log, err)
			return err
		}

		if err = tx.Model(log).Updates(map[string]interface{}{
			"status":      status,
			"gas_used":    gasUsed,
			"executed_at": executedAt,
		}).Error; err != nil {
			logrus.Errorf("set final status failed log: %+v err: %+v", log, err)
			return err
		}

		return nil
	})

	if err != nil {
		return log, err
	}

	// reload
	DB.First(log, log.ID)
	return log, nil
}

func ExecuteInRepeatableReadTransaction(callback func(tx *gorm.DB) error) (err error) {
	tryTimes := 0

	for i := 0; i < 5; i++ {
		if tryTimes != 0 {
			time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
		}

		tryTimes = tryTimes + 1

		if tryTimes > 3 {
			logrus.Errorf("tx finial failed after several retries")
			return
		}

		tx := DB.Begin()
		err = tx.Exec(`set transaction isolation level repeatable read`).Error

		if err != nil {
			tx.Rollback()
			continue
		}

		if err = callback(tx); err != nil {
			tx.Rollback()
			continue
		}

		if err = tx.Commit().Error; err != nil {
			tx.Rollback()
			logrus.Error("commit failed")
			continue
		}

		break
	}

	return
}
