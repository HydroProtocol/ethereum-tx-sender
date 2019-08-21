package main

import (
	"context"
	"github.com/HydroProtocol/nights-watch"
	"github.com/HydroProtocol/nights-watch/plugin"
	"github.com/HydroProtocol/nights-watch/structs"
	"github.com/sirupsen/logrus"
)

func startNightWatch(ctx context.Context) {
	w := nights_watch.NewHttpBasedEthWatcher(ctx, config.EthereumNodeUrl)

	w.RegisterTxReceiptPlugin(plugin.NewTxReceiptPlugin(func(txAndReceipt *structs.RemovableTxAndReceipt) {
		if txAndReceipt.IsRemoved {
			return
		}

		var log LaunchLog
		db.Where("hash = ?", txAndReceipt.Receipt.GetTxHash()).First(&log)

		if log.ID == 0 && log.From == "" {
			return
		}

		var result string
		var err error
		gasUsed := 0
		executedAt := 0

		// to get gasUsed, TODO: return gasUsed from receipt
		receipt, err := ethrpcClient.EthGetTransactionReceipt(txAndReceipt.Receipt.GetTxHash())

		if err != nil || receipt == nil || receipt.TransactionHash == "" {
			logrus.Errorf("get receipt gasUsed failed, err: %+v", err)
		} else {
			gasUsed = receipt.GasUsed
		}

		block, err := ethrpcClient.EthGetBlockByNumber(receipt.BlockNumber, false)

		if err != nil {
			logrus.Errorf("get receipt block timestamp failed, err: %+v", err)
		} else {
			executedAt = block.Timestamp
		}

		if txAndReceipt.Receipt.GetResult() {
			result = "successful"
			err = handleLaunchLogStatus(&log, true, gasUsed, executedAt)
		} else {
			result = "failed"
			err = handleLaunchLogStatus(&log, false, gasUsed, executedAt)
		}

		logrus.Infof("tx %s err: %+v result: %s", txAndReceipt.Receipt.GetTxHash(), err, result)
	}))

	_ = w.RunTillExit()
}
