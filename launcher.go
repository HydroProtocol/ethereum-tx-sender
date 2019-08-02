package main

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	pb "git.ddex.io/infrastructure/ethereum-launcher/messages"
	"git.ddex.io/lib/ethrpc"
	"git.ddex.io/lib/monitor"
	"github.com/davecgh/go-spew/spew"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"math"
	"math/big"
	"math/rand"
	"sort"
	"time"
)

// Encode encodes b as a hex string with 0x prefix.
func Encode(b []byte) string {
	enc := make([]byte, len(b)*2+2)
	copy(enc, "0x")
	hex.Encode(enc[2:], b)
	return string(enc)
}

func decimalToBigInt(d decimal.Decimal) *big.Int {
	n := new(big.Int)
	n, ok := n.SetString(d.String(), 10)
	if !ok {
		logrus.Fatalf("decimal to big int failed d: %s", d.String())
	}
	return n
}

func executeInRepeatableReadTransaction(callback func(tx *gorm.DB) error) error {
	tryTimes := 0

	for i := 0; i < 5; i++ {
		if tryTimes != 0 {
			time.Sleep(time.Duration(rand.Intn(300)) * time.Millisecond)
		}

		tryTimes = tryTimes + 1

		if tryTimes > 3 {
			logrus.Errorf("tx finial failed after several retries")
			return fmt.Errorf("tx finial failed after several retries")
		}

		tx := db.Begin()
		err := tx.Exec(`set transaction isolation level repeatable read`).Error

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

	return nil
}

func handleLaunchLogStatus(log *LaunchLog, result bool) error {
	log.Status = pb.LaunchLogStatus_name[int32(pb.LaunchLogStatus_SUCCESS)]
	var status string

	if result {
		status = pb.LaunchLogStatus_name[int32(pb.LaunchLogStatus_SUCCESS)]
	} else {
		status = pb.LaunchLogStatus_name[int32(pb.LaunchLogStatus_FAILED)]
	}

	return executeInRepeatableReadTransaction(func(tx *gorm.DB) (err error) {
		if err = tx.Model(LaunchLog{}).Where(
			"item_type = ? and item_id = ? and status = ? and hash != ?",
			log.ItemType,
			log.ItemID,
			pb.LaunchLogStatus_name[int32(pb.LaunchLogStatus_PENDING)],
			log.Hash,
		).Update(map[string]interface{}{
			"status": pb.LaunchLogStatus_name[int32(pb.LaunchLogStatus_RETRIED)],
		}).Error; err != nil {
			logrus.Errorf("set retry status failed log: %+v err: %+v", log, err)
			tx.Rollback()
			return err
		}

		if err = tx.Model(log).Update("status", status).Error; err != nil {
			logrus.Errorf("set final status failed log: %+v err: %+v", log, err)
			tx.Rollback()
			return err
		}

		return nil
	})
}

func sendEthLaunchLogWithGasPrice(launchLog *LaunchLog, gasPrice decimal.Decimal) (txHash string, err error) {
	isNewLog := true

	if launchLog.Nonce.Valid {
		isNewLog = false
	}

	var nonce uint64

	if isNewLog {
		nonce = uint64(getNextNonce(launchLog.From))
	} else {
		nonce = uint64(launchLog.Nonce.Int64)
	}

	t := ethrpc.T{
		From:     launchLog.From,
		To:       launchLog.To,
		Data:     Encode(launchLog.Data),
		Value:    decimalToBigInt(launchLog.Value),
		GasPrice: decimalToBigInt(gasPrice),
		Nonce:    int(nonce),
	}

	spew.Dump("launlog is", launchLog)
	var gasLimit uint64
	// if gas limit is empty
	// try to get gas limitation with retry times
	if launchLog.GasLimit == 0 {
		for i := 0; i < 3; i++ {
			gas, err := ethrpcClient.EthEstimateGas(t)
			spew.Dump(gas, err)
			if err != nil {
				continue
			}

			if gas == 0 {
				launchLog.Status = pb.LaunchLogStatus_name[int32(pb.LaunchLogStatus_SIGN_FAILED)]
				return "", fmt.Errorf("EthEstimateGas = 0")
			} else {
				logrus.Info("EthEstimateGas for %d is %s", launchLog.ID, gas)
			}

			gasLimit = uint64(float64(gas) * 1.2)
			launchLog.GasLimit = gasLimit
			break
		}

		if err != nil {
			return "", err
		}
	} else {
		gasLimit = launchLog.GasLimit
	}

	spew.Dump("gasLimit", gasLimit)

	t.Gas = int(gasLimit)

	rawTxHex, err := pkmSign(&t)

	if err != nil {
		launchLog.Status = pb.LaunchLogStatus_name[int32(pb.LaunchLogStatus_SIGN_FAILED)]
		return "", err
	}

	hash, err := ethrpcClient.EthSendRawTransaction(rawTxHex)

	if err != nil {
		return "", err
	}

	launchLog.Hash = sql.NullString{
		String: hash,
		Valid:  true,
	}

	launchLog.GasPrice = gasPrice

	// only inc if isNewLaunchLog
	// otherwise it is resend, keep the nonce
	if isNewLog {
		launchLog.Nonce = sql.NullInt64{
			Int64: int64(nonce),
			Valid: true,
		}

		increaseNextNonce(launchLog.From)
	}

	launchLog.Status = pb.LaunchLogStatus_name[int32(pb.LaunchLogStatus_PENDING)]
	logrus.Infof("send launcher log, hash: %s, rawTxString: %s", hash, rawTxHex)

	return hash, err

}

func StartSendLoop(ctx context.Context) {
	logrus.Info("send loop start!")

	for {
		launchLogs := getAllLogsWithStatus(pb.LaunchLogStatus_name[int32(pb.LaunchLogStatus_CREATED)])

		if len(launchLogs) == 0 {
			select {
			case <-ctx.Done():
				logrus.Info("launcher send loop Exit")
				return
			case <-time.After(5 * time.Second):
				logrus.Infof("no logs need to be sent. sleep 5s")
				continue
			}
		}

		logrus.Infof("%d created log to be send", len(launchLogs))

		gasPrice := getCurrentGasPrice()

		for i := 0; i < len(launchLogs); i++ {
			start := time.Now()

			launchLog := launchLogs[i]

			_, err := sendEthLaunchLogWithGasPrice(launchLog, gasPrice)

			if err != nil {
				monitor.Count("launcher_shoot_failed")
				logrus.Errorf("shoot launch log error id %d, err %v, err msg: %s", launchLog.ID, err, err.Error())

				// if it's signature error, do not panic
				// then the launch log will be saved for further investigate
				if launchLog.Status != pb.LaunchLogStatus_name[int32(pb.LaunchLogStatus_SIGN_FAILED)] {
					panic(err)
				}
			}

			if err = db.Save(launchLog).Error; err != nil {
				monitor.Count("launcher_update_failed")
				logrus.Errorf("update launch log error id %d, err %v", launchLog.ID, err)
				panic(err)
			}

			monitor.Time("launcher_send_log", float64(time.Since(start))/1000000)
		}
	}
}

func StartRetryLoop(ctx context.Context) {
	logrus.Info("retry loop start!")

	pendingStatusName := pb.LaunchLogStatus_name[int32(pb.LaunchLogStatus_PENDING)]

	for {
		launchLogs := getAllLogsWithStatus(pendingStatusName)
		longestPendingSecs := getLongestPendingSeconds(launchLogs)

		latestLogsForEachNonce := pickLatestLogForEachNonce(launchLogs)
		needResendLogs := pickLaunchLogsPendingTooLong(latestLogsForEachNonce)

		if len(needResendLogs) <= 0 {
			logrus.Info("no logs need to be retried. sleep 10s")
			<-time.After(10 * time.Second)
			continue
		}

		logrus.Infof("resending long pending logs, num: %d", len(needResendLogs))

		for _, launchLog := range needResendLogs {
			start := time.Now()

			gasPrice := determineGasPriceForRetryLaunchLog(launchLog, longestPendingSecs)

			_, err := sendEthLaunchLogWithGasPrice(launchLog, gasPrice)

			if err != nil {
				monitor.Count("launcher_shoot_retry_failed")
				logrus.Errorf("shoot retry launch log error id %d, err %v", launchLog.ID, err)
				panic(err)
			}

			isNewLaunchLogCreated := false

			err = executeInRepeatableReadTransaction(func(tx *gorm.DB) (er error) {
				// optimistic lock the retried launchlog
				var reloadedLog LaunchLog
				if er = tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", launchLog.ID).Select(&reloadedLog).Error; er != nil {
					return er
				}

				// if the log is no longer a pending status, skip the retry
				if reloadedLog.Status != pendingStatusName {
					return nil
				}

				if er = insertRetryLaunchLog(tx, launchLog); er != nil {
					return er
				}

				isNewLaunchLogCreated = true

				return nil
			})

			if err != nil {
				monitor.Count("launcher_retry_insert_failed")
				logrus.Errorf("insert launch log error id %d, err %v", launchLog.ID, err)
				panic(err)
			}

			if isNewLaunchLogCreated {
				monitor.Count("launcher_retry_count")
				monitor.Time("launcher_retry", float64(time.Since(start))/1000000)

				gasPriceInGwei, _ := gasPrice.Div(decimal.New(1, 9)).Float64()
				monitor.Value("launcher_retry_gas_price_in_gwei", gasPriceInGwei)
			}
		}

		logrus.Infoln("done resending long pending logs")
	}

}

func determineGasPriceForRetryLaunchLog(launchLog *LaunchLog, longestPendingSecs int) decimal.Decimal {
	suggestGasPrice := getCurrentGasPrice()

	minRetryGasPrice := launchLog.GasPrice.Mul(decimal.New(115, -2))
	gasPrice := decimal.Max(suggestGasPrice, minRetryGasPrice)
	increasedGasPrice := increaseGasPriceAccordingToPendingTime(longestPendingSecs, gasPrice)

	maxGasPrice := config.MaxGasPriceForRetry
	determinedPrice := decimal.Min(increasedGasPrice, maxGasPrice)
	logrus.Infof("gas price for retry launch log(nonce: %v), suggest: %s, minRetry: %s, increasedGasPrice: %s, final: %s", launchLog.Nonce, suggestGasPrice, minRetryGasPrice, increasedGasPrice, determinedPrice)

	return determinedPrice.Round(0)
}

func increaseGasPriceAccordingToPendingTime(pendingSeconds int, gasPrice decimal.Decimal) decimal.Decimal {
	// after subtract 2 minutes, for every extra minute, we add 10%
	increaseRatio := 0.1 * (math.Max(float64(pendingSeconds-2*60), 0) / 60)
	logrus.Infoln("increaseGasPriceAccordingToPendingTime ratio:", increaseRatio)
	gasAfterIncrease := gasPrice.Mul(decimal.NewFromFloat(1 + increaseRatio))

	return gasAfterIncrease
}

func getLongestPendingSeconds(logs []*LaunchLog) int {
	pendingSeconds := 0

	for _, log := range logs {
		pendingDuration := time.Now().Sub(log.CreatedAt)
		curPendingSecs := int(pendingDuration.Seconds())

		if curPendingSecs > pendingSeconds {
			pendingSeconds = curPendingSecs
		}
	}

	return pendingSeconds
}

func pickLatestLogForEachNonce(logs []*LaunchLog) (rst []*LaunchLog) {
	// nonce -> latest launcher log
	holderMap := make(map[string]*LaunchLog)

	for _, log := range logs {
		key := fmt.Sprintf("%s-%d", log.From, log.Nonce.Int64)
		if existValue, exist := holderMap[key]; exist {
			if log.CreatedAt.After(existValue.CreatedAt) {
				holderMap[key] = log
			}
		} else {
			holderMap[key] = log
		}
	}

	for _, v := range holderMap {
		rst = append(rst, v)
	}

	// sort by nonce, progressive
	sort.Slice(rst, func(i, j int) bool {
		return rst[i].Nonce.Int64 < rst[j].Nonce.Int64
	})

	return
}

func pickLaunchLogsPendingTooLong(logs []*LaunchLog) (rst []*LaunchLog) {
	timeoutForLaunchlogPendingInSecs := config.RetryPendingSecondsThreshold

	for _, launchLog := range logs {
		gapBackward := time.Duration(-1*timeoutForLaunchlogPendingInSecs) * time.Second
		oldBoundaryLine := time.Now().Add(gapBackward).UTC()

		tooOld := launchLog.CreatedAt.Before(oldBoundaryLine)
		if tooOld {
			rst = append(rst, launchLog)
		}
	}

	return
}

func insertRetryLaunchLog(tx *gorm.DB, launchLog *LaunchLog) error {
	newLog := &LaunchLog{
		ItemType: launchLog.ItemType,
		ItemID:   launchLog.ItemID,
		Status:   pb.LaunchLogStatus_name[int32(pb.LaunchLogStatus_PENDING)],
		From:     launchLog.From,
		To:       launchLog.To,
		Value:    launchLog.Value,
		GasLimit: launchLog.GasLimit,
		Data:     launchLog.Data,
		Nonce:    launchLog.Nonce,
		Hash: sql.NullString{
			String: launchLog.Hash.String,
			Valid:  true,
		},
		GasPrice: launchLog.GasPrice,
	}

	if err := tx.Save(newLog).Error; err != nil {
		return err
	}

	// TODO use subscribe instead
	// err = updateTransactionAndTrades(newLog)

	return nil
}

func StartLauncher(ctx context.Context) {
	go StartRetryLoop(ctx)
	StartSendLoop(ctx)
}
