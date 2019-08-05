package main

import (
	"os"
	"testing"
)

func TestRun(t *testing.T) {
	_ = os.Setenv("KUBE_NAMESPACE", "sandbox")
	_ = os.Setenv("KUBE_APP_NAME", "ethereum-launcher-kovan")
	_ = os.Setenv("ETCD_SERVER_ADDRESS", "http://0.0.0.0:8888")

	config = &Config{
		DatabaseURL: "postgres://david:@localhost:5432/launcher",
	}

	connectDB()

	var reloadedLog LaunchLog
	db.Model(&reloadedLog).Debug().Set("gorm:query_option", "FOR UPDATE").Where("id = ?", 1).Scan(&reloadedLog)
	//run()
}
