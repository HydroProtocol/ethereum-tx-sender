package main

import (
	"context"
	"git.ddex.io/infrastructure/ethereum-launcher/api"
	"git.ddex.io/infrastructure/ethereum-launcher/config"
	"git.ddex.io/infrastructure/ethereum-launcher/launcher"
	"git.ddex.io/infrastructure/ethereum-launcher/models"
	"git.ddex.io/infrastructure/ethereum-launcher/watcher"
	"git.ddex.io/lib/ethrpc"
	"git.ddex.io/lib/log"
	"git.ddex.io/lib/monitor"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	os.Exit(run())
}

func run() int {
	configs,_ := config.InitConfig()
	ctx, stop := context.WithCancel(context.Background())

	logrus.Infof("config is: %+v", configs)

	ethrpcClient := ethrpc.New(configs.EthereumNodeUrl)

	log.AutoSetLogLevel()

	models.ConnectDB(configs.DatabaseURL)
	defer models.DB.Close()

	go waitExitSignal(stop)
	go monitor.StartMonitorHttpServer(ctx)

	watcherClient := watcher.NewWatcher(ctx, configs.EthereumNodeUrl, ethrpcClient)
	go watcherClient.StartWatcher()

	go api.StartGRPCServer(ctx)
	go api.StartHTTPServer(ctx)
	launcher.StartLauncher(ctx)

	return 0
}

func waitExitSignal(ctxStop context.CancelFunc) {
	var exitSignal = make(chan os.Signal)
	signal.Notify(exitSignal, syscall.SIGTERM)
	signal.Notify(exitSignal, syscall.SIGINT)

	<-exitSignal
	logrus.Info("Stopping...")
	ctxStop()
}
