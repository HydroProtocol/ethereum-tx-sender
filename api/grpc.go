package api

import (
	"context"
	"fmt"
	pb "git.ddex.io/infrastructure/ethereum-launcher/messages"
	"git.ddex.io/infrastructure/ethereum-launcher/models"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"io"
	"net"
)
//go:generate protoc -I.  --go_out=plugins=grpc:. ./messages/messages.proto

type server struct{}

func (*server) Create(ctx context.Context, msg *pb.CreateMessage) (*pb.CreateReply, error) {
	return createLog(msg)
}

type CreateCallbackFunc func(l *models.LaunchLog, err error)

func (*server) Hello(ctx context.Context, msg *pb.HelloMessage) (*pb.HelloReply, error) {
	return &pb.HelloReply{}, nil
}

func (*server) Get(ctx context.Context, msg *pb.GetMessage) (*pb.GetReply, error) {
	return getLog(msg)
}

// the launcher has it's own watcher, no need to notify
func (*server) Notify(ctx context.Context, msg *pb.NotifyMessage) (*pb.NotifyReply, error) {
	return nil, fmt.Errorf("no implement")
}

func (*server) Subscribe(subscribeServer pb.Launcher_SubscribeServer) error {
	for {
		in, err := subscribeServer.Recv()

		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		if in.Hash == "" && in.ItemType == "" && in.ItemId == "" {
			return fmt.Errorf("need at lease (hash) or (itemType + itemId) needs to be provided")
		}

		key := getSubscribeHubKey(in.ItemType, in.ItemId)
		logrus.Printf("Received Subscribe value, key: %s", key)

		subscribeHub.Register(key, subscribeServer)
		defer subscribeHub.Remove(key, subscribeServer)
	}
}

func StartGRPCServer(ctx context.Context) {
	lis, err := net.Listen("tcp", ":3001")

	if err != nil {
		logrus.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterLauncherServer(s, &server{})

	logrus.Info("gRPC endpoint is listening on 0.0.0.0:3001\n")

	if err := s.Serve(lis); err != nil {
		logrus.Fatalf("failed to serve: %v", err)
	}
}
