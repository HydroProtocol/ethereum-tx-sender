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

		if txAndReceipt.Receipt.GetResult() {
			result = "successful"
			err = handleLaunchLogStatus(&log, true)
		} else {
			result = "failed"
			err = handleLaunchLogStatus(&log, false)
		}

		logrus.Infof("tx %s err: %+v result: %s", txAndReceipt.Receipt.GetTxHash(), err, result)
	}))

	_ = w.RunTillExit()
}
