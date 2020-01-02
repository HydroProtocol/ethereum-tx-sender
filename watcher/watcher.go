package watcher

import (
	"context"
	"fmt"
	"git.ddex.io/infrastructure/ethereum-launcher/api"
	"git.ddex.io/infrastructure/ethereum-launcher/models"
	"github.com/onrik/ethrpc"
	"git.ddex.io/lib/monitor"
	"github.com/HydroProtocol/nights-watch"
	"github.com/HydroProtocol/nights-watch/plugin"
	"github.com/HydroProtocol/nights-watch/structs"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Watcher struct {
	lastSavedBlockNumber   int
	updateBlockNumberMutex *sync.Mutex
	ethereumNodeUrl        string
	ethrpcClient           *ethrpc.EthRPC
	ctx                    context.Context
}

func NewWatcher(ctx context.Context, ethNodeUrl string, ethrpcClient *ethrpc.EthRPC) *Watcher {

	lastSavedBlockNumber, err := models.BlockNumberDao.GetCurrentBlockNumber()
	if err != nil {
		panic(err)
	}

	return &Watcher{
		ctx:                    ctx,
		lastSavedBlockNumber:   lastSavedBlockNumber,
		updateBlockNumberMutex: &sync.Mutex{},
		ethereumNodeUrl:        ethNodeUrl,
		ethrpcClient:           ethrpcClient,
	}
}

func (w *Watcher) StartWatcher() {
	nightWatch := nights_watch.NewHttpBasedEthWatcher(w.ctx, w.ethereumNodeUrl)

	nightWatch.RegisterTxReceiptPlugin(plugin.NewTxReceiptPlugin(func(txAndReceipt *structs.RemovableTxAndReceipt) {
		if txAndReceipt.IsRemoved {
			return
		}

		_ = w.saveBlockNumber(int(txAndReceipt.Receipt.GetBlockNumber()))
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
		receipt, err := w.ethrpcClient.EthGetTransactionReceipt(txAndReceipt.Receipt.GetTxHash())

		if err != nil || receipt == nil || receipt.TransactionHash == "" {
			logrus.Errorf("get receipt gasUsed failed, err: %+v, receipt: %+v", err, receipt)
		} else {
			gasUsed = receipt.GasUsed
		}

		var handledLog *models.LaunchLog
		if txAndReceipt.Receipt.GetResult() {
			result = "successful"
			handledLog, err = models.HandleLaunchLogStatus(log, true, gasUsed, executedAt)
		} else {
			result = "failed"
			handledLog, err = models.HandleLaunchLogStatus(log, false, gasUsed, executedAt)
		}

		api.SendLogStatusToSubscriber(handledLog, err)

		logrus.Infof("tx %s err: %+v result: %s", txAndReceipt.Receipt.GetTxHash(), err, result)
	}))

	for {
		err := nightWatch.RunTillExitFromBlock(uint64(getHighestSyncedBlock()))

		if err != nil {
			logrus.Errorf("watcher error: %+v", err)
			time.Sleep(1 * time.Second)
		}
	}
}

func (w *Watcher) saveBlockNumber(blockNum int) error {
	w.updateBlockNumberMutex.Lock()
	defer w.updateBlockNumberMutex.Unlock()

	if blockNum <= w.lastSavedBlockNumber {
		return nil
	}

	_, err := models.BlockNumberDao.IncreaseBlockNumber(blockNum)

	if err != nil {
		logrus.Warnf("save block number %d fail", blockNum)
	} else {
		w.lastSavedBlockNumber = blockNum
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
