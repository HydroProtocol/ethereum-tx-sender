package main

import (
	"context"
	"git.ddex.io/infrastructure/ethereum-launcher/api"
	"git.ddex.io/infrastructure/ethereum-launcher/config"
	"git.ddex.io/infrastructure/ethereum-launcher/launcher"
	"git.ddex.io/infrastructure/ethereum-launcher/models"
	"git.ddex.io/infrastructure/ethereum-launcher/watcher"
	_ "github.com/joho/godotenv/autoload"
	"github.com/onrik/ethrpc"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	os.Exit(run())
}

func run() int {
	//logrus.SetLevel(logrus.DebugLevel)

	configs, err := config.InitConfig()
	if err != nil {
		panic(err)
	}
	ctx, stop := context.WithCancel(context.Background())
	logrus.Infof("config is: %+v", configs)

	ethrpcClient := ethrpc.New(configs.EthereumNodeUrl)
	models.ConnectDB(configs.DatabaseURL)
	defer models.DB.Close()

	go waitExitSignal(stop)

	watcherClient := watcher.NewWatcher(ctx, configs.EthereumNodeUrl, ethrpcClient)
	go watcherClient.StartWatcher()

	api.InitSubscribeHub()
	go api.StartGRPCServer(ctx)
	go api.StartHTTPServer(ctx)
	launcher.StartLauncher(ctx, ethrpcClient)

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
