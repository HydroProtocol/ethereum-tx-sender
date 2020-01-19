package main

import (
	"context"
	"fmt"
	"git.ddex.io/infrastructure/ethereum-launcher/internal/api"
	"git.ddex.io/infrastructure/ethereum-launcher/internal/config"
	"git.ddex.io/infrastructure/ethereum-launcher/internal/launcher"
	"git.ddex.io/infrastructure/ethereum-launcher/internal/models"
	"git.ddex.io/infrastructure/ethereum-launcher/internal/watcher"
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
	logrus.SetLevel(logrus.InfoLevel)

	configs, err := config.InitConfig()
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("for details see https://github.com/HydroProtocol/ethereum-sender/blob/master/docs/envs.md")
		return 0
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
