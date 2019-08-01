package main

import (
	"os"
	"testing"
)

func TestRun(t *testing.T) {
	_ = os.Setenv("KUBE_NAMESPACE", "sandbox")
	_ = os.Setenv("KUBE_APP_NAME", "ethereum-launcher-kovan")
	_ = os.Setenv("ETCD_SERVER_ADDRESS", "http://0.0.0.0:8888")
	run()
}
