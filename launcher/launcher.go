package launcher

import (
	"context"
	"database/sql"
	"fmt"
	"git.ddex.io/infrastructure/ethereum-launcher/api"
	"git.ddex.io/infrastructure/ethereum-launcher/config"
	"git.ddex.io/infrastructure/ethereum-launcher/gas"
	pb "git.ddex.io/infrastructure/ethereum-launcher/messages"
	"git.ddex.io/infrastructure/ethereum-launcher/models"
	"git.ddex.io/infrastructure/ethereum-launcher/pkm"
	"git.ddex.io/infrastructure/ethereum-launcher/utils"
	"github.com/jinzhu/gorm"
	"github.com/onrik/ethrpc"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type launcher struct {
	ethrpcClient *ethrpc.EthRPC
	pkm pkm.Pkm
	nonceCache map[string]int64
	nonceCacheMutex *sync.Mutex
}

func (l*launcher)StartSendLoop(ctx context.Context) {
	logrus.Info("send loop start!")

	for {
		launchLogs := models.LaunchLogDao.GetAllLogsWithStatus(pb.LaunchLogStatus_CREATED.String())

		if len(launchLogs) == 0 {
			select {
			case <-ctx.Done():
				logrus.Info("launcher send loop Exit")
				return
			case <-time.After(10 * time.Second):
				logrus.Infof("no logs need to be sent. sleep 10s")
				continue
			case <-api.NewRequestChannel:
				// new request has come, start working!
				logrus.Info("newRequestChannel got message!")
				continue
			}
		}

		logrus.Infof("%d created log to be send", len(launchLogs))

		normalGasPrice,urgentGasPrice := gas.GetCurrentGasPrice()

		for i := 0; i < len(launchLogs); i++ {
			launchLog := launchLogs[i]

			if launchLog.Hash.Valid {
				if ok := l.tryLoadLaunchLogReceipt(launchLog); ok {
					continue
				}
			}

			var err error
			if launchLog.IsUrgent {
				_, err = l.sendEthLaunchLogWithGasPrice(launchLog, urgentGasPrice)
			} else {
				_, err = l.sendEthLaunchLogWithGasPrice(launchLog, normalGasPrice)
			}

			if err != nil {
				logrus.Errorf("shoot launch log error id %d, err %v, err msg: %s", launchLog.ID, err, err.Error())

				if strings.Contains(strings.ToLower(err.Error()), "nonce too low") {
					l.deleteCachedNonce(launchLog.From)
					continue
				} else if strings.Contains(strings.ToLower(err.Error()), "insufficient funds") {
					launchLog.Status = pb.LaunchLogStatus_SEND_FAILED.String()
					launchLog.ErrMsg = err.Error()
					launchLog.Hash = sql.NullString{}
				} else if strings.Contains(err.Error(), "estimate gas error") {
					launchLog.Status = pb.LaunchLogStatus_ESTIMATED_GAS_FAILED.String()
					launchLog.ErrMsg = err.Error()
				} else if strings.Contains(err.Error(), "sign error") {
					launchLog.Status = pb.LaunchLogStatus_SIGN_FAILED.String()
					launchLog.ErrMsg = err.Error()
				}
			}

			if err = models.DB.Save(launchLog).Error; err != nil {
				if strings.Contains(err.Error(), "duplicate key") && strings.Contains(err.Error(), "launch_logs_hash") {
					var l models.LaunchLog
					models.DB.Model(&models.LaunchLog{}).First(&l, "hash = ?", launchLog.Hash.String)
					logrus.Errorf("update launch log error id %d, err %v, same hash id: %d", launchLog.ID, err, l.ID)
				} else {
					logrus.Errorf("update launch log error id %d, err %v", launchLog.ID, err)
				}

				api.SendLogStatusToSubscriber(launchLog, err)
				panic(err)
			}

			models.DB.First(launchLog, launchLog.ID)
			api.SendLogStatusToSubscriber(launchLog, nil)
		}
	}
}

func (l*launcher)StartRetryLoop(ctx context.Context) {
	logrus.Info("retry loop start!")

	pendingStatusName := pb.LaunchLogStatus_PENDING.String()

	for {
		launchLogs := models.LaunchLogDao.GetAllLogsWithStatus(pendingStatusName)
		longestPendingSecs := getLongestPendingSeconds(launchLogs)

		latestLogsForEachNonce := pickLatestLogForEachNonce(launchLogs)
		needResendLogs := pickLaunchLogsPendingTooLong(latestLogsForEachNonce)

		if len(needResendLogs) <= 0 {
			select {
			case <-ctx.Done():
				logrus.Info("launcher retry loop Exit")
				return
			case <-time.After(10 * time.Second):
				logrus.Info("no logs need to be retried. sleep 10s")
				continue
			}
		}

		idxOfLastUrgentNeedResendLog := -1
		for i, launchLog := range needResendLogs {
			if launchLog.IsUrgent {
				idxOfLastUrgentNeedResendLog = i
			}
		}

		logrus.Infof("resending long pending logs, num: %d", len(needResendLogs))
		var err error
		for i, launchLog := range needResendLogs {
			// try to load launch log before retry
			if ok := l.tryLoadLaunchLogReceipt(launchLog); ok {
				continue
			}

			isBlockingUrgentLog := i <= idxOfLastUrgentNeedResendLog
			if isBlockingUrgentLog {
				logrus.Infof("is blocking urgent, %d(%d) <= %d(%d)",
					i, needResendLogs[i].ID, idxOfLastUrgentNeedResendLog, needResendLogs[idxOfLastUrgentNeedResendLog].ID)
			}

			gasPrice := determineGasPriceForRetryLaunchLog(launchLog, longestPendingSecs, isBlockingUrgentLog)

			if gasPrice.Equal(launchLog.GasPrice) {
				logrus.Infof("Retry gas Price is same, skip ID: %d", launchLog.ID)
				continue
			}

			isNewLaunchLogCreated := false

			err = models.ExecuteInRepeatableReadTransaction(func(tx *gorm.DB) (er error) {
				// optimistic lock the retried launchlog g
				var reloadedLog models.LaunchLog
				if er = tx.Model(&reloadedLog).Set("gorm:query_option", "FOR UPDATE").Where("id = ?", launchLog.ID).Scan(&reloadedLog).Error; er != nil {
					return er
				}

				// if the log is no longer a pending status, skip the retry
				if reloadedLog.Status != pendingStatusName {
					return nil
				}

				// This update is important
				// We have to make some changes to fail other concurrent transactions
				if er := tx.Model(&reloadedLog).Update("updated_at", time.Now().Unix()).Error; er != nil {
					return er
				}

				_, er = l.sendEthLaunchLogWithGasPrice(launchLog, gasPrice)

				if er != nil && strings.Contains(er.Error(), "nonce too low") {
					// It means one of the tx with this nonce is finalized. Skip...
					logrus.Infof("launch_log retry return nonce too low. skip id: %d", launchLog.ID)
					return nil
				}

				if er != nil {
					logrus.Infof("sendEthLaunchLogWithGasPrice() failed, sendEthLaunchLogWithGasPrice(id: %d, gasPrice: %s), err: %s",
						launchLog.ID, gasPrice, er)
					return er
				}

				if er = models.LaunchLogDao.InsertRetryLaunchLog(tx, launchLog); er != nil {
					return er
				}

				isNewLaunchLogCreated = true

				return nil
			})

			if err != nil {
				logrus.Errorf("insert launch log error id %d, err %v", launchLog.ID, err)
				panic(err)
			}
		}

		logrus.Infoln("done resending long pending logs")
	}
}

func (l*launcher)sendEthLaunchLogWithGasPrice(launchLog *models.LaunchLog, gasPrice decimal.Decimal) (txHash string, err error) {
	isNewLog := true

	if launchLog.Nonce.Valid {
		isNewLog = false
	}

	var nonce uint64

	if isNewLog {
		nonce = uint64(l.getNextNonce(launchLog.From))
	} else {
		nonce = uint64(launchLog.Nonce.Int64)
	}

	t := ethrpc.T{
		From:     launchLog.From,
		To:       launchLog.To,
		Data:     utils.Encode(launchLog.Data),
		Value:    utils.DecimalToBigInt(launchLog.Value),
		GasPrice: utils.DecimalToBigInt(gasPrice),
		Nonce:    int(nonce),
	}

	var gasLimit uint64
	// if gas limit is empty
	// try to get gas limitation with retry times
	if launchLog.GasLimit == 0 {
		for i := 0; i < 2; i++ {
			var estimateGas int
			estimateGas, err = l.ethrpcClient.EthEstimateGas(t)

			if err != nil {
				continue
			}

			gasLimit = uint64(float64(estimateGas) * 1.2)
			launchLog.GasLimit = gasLimit
			break
		}

		if err != nil {
			return "", fmt.Errorf("estimate gas error %+v", err)
		}
	} else {
		gasLimit = launchLog.GasLimit
	}

	t.Gas = int(gasLimit)

	rawTxHex, err := l.pkm.Sign(&t)

	if err != nil {
		return "", fmt.Errorf("sign error %+v", err)
	}

	hash := utils.EncodeHex(utils.Keccak256(utils.DecodeHex(rawTxHex)))

	launchLog.Hash = sql.NullString{
		String: hash,
		Valid:  true,
	}

	hashOnChain, err := l.ethrpcClient.EthSendRawTransaction(rawTxHex)

	if err != nil {
		return "", err
	}

	if hashOnChain != hash {
		logrus.Fatalf("hashOnChain != hash, %s, %s", hashOnChain, hash)
	} else {
		logrus.Infof("send tx hash: %s, isNewLog: %t", hash, isNewLog)
	}

	launchLog.GasPrice = gasPrice

	// only inc if isNewLaunchLog
	// otherwise it is resend, keep the nonce
	if isNewLog {
		launchLog.Nonce = sql.NullInt64{
			Int64: int64(nonce),
			Valid: true,
		}

		l.increaseNextNonce(launchLog.From)
	}

	launchLog.Status = pb.LaunchLogStatus_PENDING.String()
	logrus.Infof("send launcher log, hash: %s, rawTxString: %s", hash, rawTxHex)

	return hash, err
}

func (l*launcher)tryLoadLaunchLogReceipt(launchLog *models.LaunchLog) bool {
	receipt, err := l.ethrpcClient.EthGetTransactionReceipt(launchLog.Hash.String)

	if err != nil || receipt == nil || receipt.TransactionHash == "" {
		return false
	}

	var result string
	status, _ := strconv.ParseInt(receipt.Status, 0, 0)

	gasUsed := receipt.GasUsed

	block, err := l.ethrpcClient.EthGetBlockByNumber(receipt.BlockNumber, false)

	if err != nil {
		return false
	}

	executedAt := block.Timestamp

	var handledLog *models.LaunchLog
	if status == 1 {
		result = "successful"
		handledLog, err = models.HandleLaunchLogStatus(launchLog, true, gasUsed, executedAt)
	} else {
		result = "failed"
		handledLog, err = models.HandleLaunchLogStatus(launchLog, false, gasUsed, executedAt)
	}

	api.SendLogStatusToSubscriber(handledLog, err)
	logrus.Infof("log %s receipt request finial %s, err: %+v", launchLog.Hash.String, result, err)

	if err != nil {
		return false
	}

	return true
}

func determineGasPriceForRetryLaunchLog(
	launchLog *models.LaunchLog,
	longestPendingSecs int,
	isBlockingUrgentLog bool,
) decimal.Decimal {
	treatAsUrgent := isBlockingUrgentLog || launchLog.IsUrgent

	normalPrice, urgentPrice := gas.GetCurrentGasPrice()
	suggestGasPrice := urgentPrice
	if treatAsUrgent {
		suggestGasPrice = normalPrice
	}

	minRetryGasPrice := launchLog.GasPrice.Mul(decimal.New(115, -2))
	gasPrice := decimal.Max(suggestGasPrice, minRetryGasPrice)
	increasedGasPrice := increaseGasPriceAccordingToPendingTime(longestPendingSecs, gasPrice)

	maxGasPrice := config.Config.MaxGasPriceForRetry
	determinedPrice := decimal.Min(increasedGasPrice, maxGasPrice)
	logrus.Debugf("gas price for retry launch log(nonce: %v), suggest: %s, minRetry: %s, increasedGasPrice: %s, final: %s", launchLog.Nonce, suggestGasPrice, minRetryGasPrice, increasedGasPrice, determinedPrice)

	return determinedPrice.Round(0)
}

func increaseGasPriceAccordingToPendingTime(pendingSeconds int, gasPrice decimal.Decimal) decimal.Decimal {
	// after subtract 2 minutes, for every extra minute, we add 10%
	increaseRatio := 0.1 * (math.Max(float64(pendingSeconds-2*60), 0) / 60)
	logrus.Debugln("increaseGasPriceAccordingToPendingTime ratio:", increaseRatio)
	gasAfterIncrease := gasPrice.Mul(decimal.NewFromFloat(1 + increaseRatio))

	return gasAfterIncrease
}

func getLongestPendingSeconds(logs []*models.LaunchLog) int {
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

func pickLatestLogForEachNonce(logs []*models.LaunchLog) (rst []*models.LaunchLog) {
	// nonce -> latest launcher log
	holderMap := make(map[string]*models.LaunchLog)

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

func pickLaunchLogsPendingTooLong(logs []*models.LaunchLog) (rst []*models.LaunchLog) {
	// make sure logs are sort by nonce asc
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Nonce.Int64 < logs[j].Nonce.Int64
	})

	timeoutForLaunchlogPendingInSecs := config.Config.RetryPendingSecondsThreshold
	timeoutForUrgentLaunchlogPendingInSecs := config.Config.RetryPendingSecondsThresholdForUrgent

	// in case urgent not set
	if timeoutForUrgentLaunchlogPendingInSecs <= 0 {
		timeoutForUrgentLaunchlogPendingInSecs = timeoutForLaunchlogPendingInSecs
	}

	oldBoundaryLineIdx := -1
	for i, launchLog := range logs {

		var gapBackward time.Duration
		if launchLog.IsUrgent {
			gapBackward = time.Duration(-1*timeoutForUrgentLaunchlogPendingInSecs) * time.Second
		} else {
			gapBackward = time.Duration(-1*timeoutForLaunchlogPendingInSecs) * time.Second
		}

		oldBoundaryLine := time.Now().Add(gapBackward).UTC()
		tooOld := launchLog.CreatedAt.Before(oldBoundaryLine)

		if tooOld {
			oldBoundaryLineIdx = i
		}
	}

	if oldBoundaryLineIdx >= 0 {
		logrus.Infof("pick pending too long, %d/%d", oldBoundaryLineIdx+1, len(logs))

		return logs[0 : oldBoundaryLineIdx+1]
	}

	return []*models.LaunchLog{}
}

func StartLauncher(ctx context.Context, ethrpcClient *ethrpc.EthRPC) {
	l := launcher{
		ethrpcClient:ethrpcClient,
		pkm:pkm.InitPKM(config.Config.PrivateKeys),
		nonceCache:make(map[string]int64),
		nonceCacheMutex: &sync.Mutex{},
	}

	go l.StartRetryLoop(ctx)
	l.StartSendLoop(ctx)
}
