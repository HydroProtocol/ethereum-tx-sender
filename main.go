package main

import (
	"context"
	"git.ddex.io/lib/ethrpc"
	"git.ddex.io/lib/hotconfig"
	"git.ddex.io/lib/log"
	"git.ddex.io/lib/monitor"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

type Config struct {
	DatabaseURL                  string          `json:"database_url"`
	MaxGasPriceForRetry          decimal.Decimal `json:"max_gas_price_for_retry"`
	RetryPendingSecondsThreshold int             `json:"retry_pending_seconds_threshold"`
	EthereumNodeUrl              string          `json:"ethereum_node_url"`
	PkmUrl                       string          `json:"pkm_url"`
}

var config *Config
var ethrpcClient *ethrpc.EthRPC

func waitExitSignal(ctxStop context.CancelFunc) {
	var exitSignal = make(chan os.Signal)
	signal.Notify(exitSignal, syscall.SIGTERM)
	signal.Notify(exitSignal, syscall.SIGINT)

	<-exitSignal
	logrus.Info("Stopping...")
	ctxStop()
}

func run() int {
	config = &Config{}
	ctx, stop := context.WithCancel(context.Background())

	if os.Getenv("KUBE_NAMESPACE") != "" {
		hotconfig.Load(config, &hotconfig.Options{
			Watch:   true,
			Context: ctx,
		})
	} else {
		config = &Config{
			DatabaseURL:                  "postgres://localhost:5432/launcher",
			MaxGasPriceForRetry:          decimal.New(5, 9), // 5 Gwei
			RetryPendingSecondsThreshold: 10,                // 10 s
		}
	}

	ethrpcClient = ethrpc.New(config.EthereumNodeUrl)

	log.AutoSetLogLevel()

	connectDB()
	defer db.Close()

	go waitExitSignal(stop)
	go monitor.StartMonitorHttpServer(ctx)

	go tryLoadPendingTxStatus(ctx)
	go startNightWatch(ctx) // TODO we may need a global watcher in the feature
	go startGrpcServer(ctx)
	StartLauncher(ctx)

	return 0
}

func main() {
	os.Exit(run())
}
