package main

import (
	"context"
	"github.com/HydroProtocol/nights-watch"
	"github.com/HydroProtocol/nights-watch/plugin"
	"github.com/HydroProtocol/nights-watch/structs"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestWathcher(t *testing.T) {
	w := nights_watch.NewHttpBasedEthWatcher(context.Background(), "https://mainnet.infura.io/v3/19d753b2600445e292d54b1ef58d4df4")

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

		if txAndReceipt.Receipt.GetResult() {
			result = "successful"
			handleLaunchLogStatus(&log, true)
		} else {
			result = "failed"
			handleLaunchLogStatus(&log, false)
		}

		logrus.Infof("tx %s result: %s", txAndReceipt.Receipt.GetTxHash(), result)
	}))

	_ = w.RunTillExit()
}
