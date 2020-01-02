package main

import (
	"database/sql"
	pb "git.ddex.io/infrastructure/ethereum-launcher/messages"
	"git.ddex.io/infrastructure/ethereum-launcher/models"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"sync"
	"testing"
	"time"
)

func TestRetryAndOriginalTxSuccess(t *testing.T) {
	// init
	config = &Config{
		DatabaseURL: "postgres://david:@localhost:5432/launcher",
	}

	models.ConnectDB(config.DatabaseURL)
	models.DB.Unscoped().Delete(models.LaunchLog{}, "'1' = ?", "1")
	models.DB.Model(&models.LaunchLog{}).Create(&models.LaunchLog{
		From:     "0x0",
		To:       "0x1",
		Value:    decimal.Zero,
		GasLimit: 100000,
		Status:   "PENDING",
		GasPrice: decimal.New(1, 10),
		Data:     []byte{},
		ItemType: "T",
		ItemID:   "id",
		Hash: sql.NullString{
			String: "original",
			Valid:  true,
		},
	})

	var originalLog models.LaunchLog
	var anotherOriginalLog models.LaunchLog
	models.DB.Model(&models.LaunchLog{}).Where("item_type = ? and item_id = ?", "T", "id").Scan(&originalLog)
	models.DB.Model(&models.LaunchLog{}).Where("item_type = ? and item_id = ?", "T", "id").Scan(&anotherOriginalLog)
	models.DB.LogMode(true)
	wg := sync.WaitGroup{}
	// set status loop
	wg.Add(1)
	go func() {
		defer wg.Done()
		status := pb.LaunchLogStatus_SUCCESS.String()
		_ = executeInRepeatableReadTransaction(func(tx *gorm.DB) (err error) {
			time.Sleep(100 * time.Millisecond)
			logrus.Info("loop 1 in")
			var reloadedLog models.LaunchLog

			if err = tx.Model(&reloadedLog).Set("gorm:query_option", "FOR UPDATE").Where("id = ?", originalLog.ID).Scan(&reloadedLog).Error; err != nil {
				logrus.Info("loop 1 lock error")
				return err
			}

			logrus.Info("loop 1 lock")

			time.Sleep(300 * time.Millisecond)

			if reloadedLog.Status != pb.LaunchLogStatus_PENDING.String() {
				return nil
			}

			if err = tx.Model(models.LaunchLog{}).Where(
				"item_type = ? and item_id = ? and status = ? and hash != ?",
				originalLog.ItemType,
				originalLog.ItemID,
				pb.LaunchLogStatus_PENDING.String(),
				originalLog.Hash,
			).Update(map[string]interface{}{
				"status": pb.LaunchLogStatus_RETRIED.String(),
			}).Error; err != nil {
				logrus.Errorf("set retry status failed log: %+v err: %+v", originalLog, err)
				tx.Rollback()
				return err
			}

			if err = tx.Model(originalLog).Update("status", status).Error; err != nil {
				logrus.Errorf("set final status failed log: %+v err: %+v", originalLog, err)
				tx.Rollback()
				return err
			}

			return nil
		})
		logrus.Info("loop 1 out")
	}()

	// retry loop
	wg.Add(1)
	go func() {
		defer wg.Done()
		status := pb.LaunchLogStatus_PENDING.String()
		_ = executeInRepeatableReadTransaction(func(tx *gorm.DB) (er error) {
			time.Sleep(100 * time.Millisecond)
			logrus.Info("loop 2 in")

			// optimistic lock the retried launchlog
			var reloadedLog models.LaunchLog
			if er = tx.Model(&reloadedLog).Set("gorm:query_option", "FOR UPDATE").Where("id = ?", originalLog.ID).Scan(&reloadedLog).Error; er != nil {
				logrus.Info("loop 2 lock error 1", er)
				return er
			}

			logrus.Info("loop 2 lock")

			time.Sleep(300 * time.Millisecond)

			// if the log is no longer a pending status, skip the retry
			if reloadedLog.Status != status {
				return nil
			}

			if er := tx.Model(&reloadedLog).Update("updated_at", time.Now().Unix()).Error; er != nil {
				logrus.Info("loop 2 lock error 2", er)
				return er
			}

			// use sleep to simulate send tx
			time.Sleep(500 * time.Millisecond)
			anotherOriginalLog.Hash = sql.NullString{
				String: "retried",
				Valid:  true,
			}
			//_, er = sendEthLaunchLogWithGasPrice(&anotherOriginalLog, gasPrice)

			if er = insertRetryLaunchLog(tx, &anotherOriginalLog); er != nil {
				logrus.Info("loop 2 lock error 3")
				return er
			}

			return nil
		})
		logrus.Info("loop 2 out")
	}()

	wg.Wait()
}
