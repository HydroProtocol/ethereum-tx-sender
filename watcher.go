package main

import (
	"context"
	"fmt"
	"git.ddex.io/infrastructure/ethereum-launcher/models"
	"git.ddex.io/lib/monitor"
	"github.com/HydroProtocol/nights-watch"
	"github.com/HydroProtocol/nights-watch/plugin"
	"github.com/HydroProtocol/nights-watch/structs"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

var lastSavedBlockNumber int
var updateBlockNumberMutex *sync.Mutex

func startNightWatch(ctx context.Context) {
	w := nights_watch.NewHttpBasedEthWatcher(ctx, config.EthereumNodeUrl)

	w.RegisterTxReceiptPlugin(plugin.NewTxReceiptPlugin(func(txAndReceipt *structs.RemovableTxAndReceipt) {
		if txAndReceipt.IsRemoved {
			return
		}

		_ = saveBlockNumber(int(txAndReceipt.Receipt.GetBlockNumber()))
		monitor.Value("block_number", float64(txAndReceipt.Receipt.GetBlockNumber()))
		log := models.LaunchLogDao.FindLogByHash(txAndReceipt.Receipt.GetTxHash())

		if log.ID == 0 && log.From == "" {
			return
		}

		var result string
		var err error
		gasUsed := 0
		executedAt := int(txAndReceipt.TimeStamp)

		// to get gasUsed, TODO: return gasUsed from receipt
		receipt, err := ethrpcClient.EthGetTransactionReceipt(txAndReceipt.Receipt.GetTxHash())

		if err != nil || receipt == nil || receipt.TransactionHash == "" {
			logrus.Errorf("get receipt gasUsed failed, err: %+v, receipt: %+v", err, receipt)
		} else {
			gasUsed = receipt.GasUsed
		}

		if txAndReceipt.Receipt.GetResult() {
			result = "successful"
			err = handleLaunchLogStatus(log, true, gasUsed, executedAt)
		} else {
			result = "failed"
			err = handleLaunchLogStatus(log, false, gasUsed, executedAt)
		}

		logrus.Infof("tx %s err: %+v result: %s", txAndReceipt.Receipt.GetTxHash(), err, result)
	}))

	for {
		err := w.RunTillExitFromBlock(uint64(getHighestSyncedBlock()))

		if err != nil {
			logrus.Errorf("watcher error: %+v", err)
			time.Sleep(1 * time.Second)
		}
	}
}


func saveBlockNumber(blockNum int) error {
	updateBlockNumberMutex.Lock()
	defer updateBlockNumberMutex.Unlock()

	if blockNum <= lastSavedBlockNumber {
		return nil
	}

	_, err := models.BlockNumberDao.IncreaseBlockNumber(blockNum)

	if err != nil {
		logrus.Warnf("save block number %d fail", blockNum)
	} else {
		lastSavedBlockNumber = blockNum
		logrus.Infof("save block number %d success", blockNum)
	}

	return err
}

func getHighestSyncedBlock() int {
	blockNumber, err := models.BlockNumberDao.GetCurrentBlockNumber()

	if err != nil {
		panic(fmt.Sprintf("err when GetHighestSyncedBlock: %s", err))
	}

	return blockNumber
}
