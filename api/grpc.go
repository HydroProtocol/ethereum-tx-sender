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
	"sync"
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

func getSubscribeHubKey(itemType, itemId string) string {
	return fmt.Sprintf("Type:%s-ID:%s", itemType, itemId)
}

func SendLogStatusToSubscriber(log *models.LaunchLog, err error) {
	logrus.Infof("SendLogStatusToSubscriber for log %d", log.ID)

	key := getSubscribeHubKey(log.ItemType, log.ItemID)

	data, ok := subscribeHub.data[key]

	if !ok || data == nil {
		logrus.Infof("no subscriber handlers found for log %d", log.ID)
		return
	}

	for s, _ := range data {
		switch v := s.(type) {
		case pb.Launcher_SubscribeServer:
			logrus.Infof("SendLogStatusToSubscriber for log %d, handler: pb.Launcher_SubscribeServer", log.ID)
			_ = v.Send(&pb.SubscribeReply{
				Status:   pb.LaunchLogStatus(pb.LaunchLogStatus_value[log.Status]),
				Hash:     log.Hash.String,
				ItemId:   log.ItemID,
				ItemType: log.ItemType,
				ErrMsg:   log.ErrMsg,
			})
		case *CreateCallbackFunc:
			logrus.Infof("SendLogStatusToSubscriber for log %d, handler: *CreateCallbackFunc", log.ID)
			(*v)(log, err)
		default:
			logrus.Errorf("SendLogStatusToSubscriber for log %d, handler: unknown, %+v, %+v", log.ID, s, v)
		}
	}
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

type SubscribeHub struct {
	m    *sync.Mutex
	data map[string]map[interface{}]bool
}

func (sb *SubscribeHub) Register(key string, handler interface{}) {
	sb.m.Lock()
	defer sb.m.Unlock()

	if _, ok := sb.data[key]; !ok {
		sb.data[key] = make(map[interface{}]bool)
	}

	sb.data[key][handler] = true
}

func (sb *SubscribeHub) Remove(key string, handler interface{}) {
	sb.m.Lock()
	defer sb.m.Unlock()

	if _, ok := sb.data[key]; !ok {
		return
	}

	delete(sb.data[key], handler)

	if len(sb.data[key]) == 0 {
		delete(sb.data, key)
	}
}

var subscribeHub *SubscribeHub

func StartGRPCServer(ctx context.Context) {
	subscribeHub = &SubscribeHub{
		m:    &sync.Mutex{},
		data: make(map[string]map[interface{}]bool),
	}

	lis, err := net.Listen("tcp", ":3000")

	if err != nil {
		logrus.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterLauncherServer(s, &server{})

	logrus.Info("gRPC endpoint is listening on 0.0.0.0:3000\n")

	if err := s.Serve(lis); err != nil {
		logrus.Fatalf("failed to serve: %v", err)
	}
}
