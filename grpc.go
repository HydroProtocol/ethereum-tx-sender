package main

import (
	"context"
	"database/sql"
	"fmt"
	pb "git.ddex.io/infrastructure/ethereum-launcher/messages"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"io"
	"net"
	"strings"
	"sync"
)

//go:generate protoc -I.  --go_out=plugins=grpc:. ./messages/messages.proto

type server struct{}

func (*server) Create(ctx context.Context, msg *pb.CreateMessage) (*pb.CreateReply, error) {
	var err error
	var value decimal.Decimal

	if msg.Value == "" {
		value = decimal.Zero
	} else {
		value, err = decimal.NewFromString(msg.Value)

		if err != nil {
			return nil, fmt.Errorf("convert value to decimal failed")
		}
	}

	var gasPrice decimal.Decimal

	if msg.GasPrice == "" {
		gasPrice = getCurrentGasPrice()
	} else {
		gasPrice, err = decimal.NewFromString(msg.GasPrice)
		if err != nil {
			return nil, fmt.Errorf("convert gas price to decimal failed")
		}
	}

	if msg.From[:2] != "0x" || len(msg.From) != 42 {
		return nil, fmt.Errorf("`form` format error, not a valid ethereum address")
	}

	if msg.To[:2] != "0x" || len(msg.To) != 42 {
		return nil, fmt.Errorf("`to` format error, not a valid ethereum address")
	}

	var count int
	if err := db.Model(&LaunchLog{}).Where("item_type = ? and item_id = ?", msg.ItemType, msg.ItemId).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("get item_type and item_id count error %v", err)
	}

	if count > 0 {
		return nil, fmt.Errorf("item_type and item_id exists !!")
	}

	log := &LaunchLog{
		Hash: sql.NullString{
			Valid: false,
		},
		From:     strings.ToLower(msg.From),
		To:       strings.ToLower(msg.To),
		Value:    value,
		GasPrice: gasPrice,
		GasLimit: uint64(msg.GasLimit),
		Nonce:    sql.NullInt64{},
		Data:     msg.Data,
		ItemID:   msg.ItemId,
		ItemType: msg.ItemType,
		Status:   pb.LaunchLogStatus_name[int32(pb.LaunchLogStatus_CREATED)],
	}

	if err = db.Create(log).Error; err != nil {
		return nil, err
	}

	return &pb.CreateReply{
		Status: pb.RequestStatus_REQUEST_SUCCESSFUL,
		ErrMsg: "",
	}, nil
}

func (*server) Hello(ctx context.Context, msg *pb.HelloMessage) (*pb.HelloReply, error) {
	return &pb.HelloReply{}, nil
}

func (*server) Get(ctx context.Context, msg *pb.GetMessage) (*pb.GetReply, error) {
	var logs []*LaunchLog

	if msg.Hash != "" {
		db.Where("hash = ?", msg.Hash).Find(&logs)
	} else if msg.ItemType != "" && msg.ItemId != "" {
		db.Where("item_type = ? and item_id = ?", msg.ItemType, msg.ItemId).Find(&logs)
	} else {
		return nil, fmt.Errorf("Need hash or (item_type, item_id) msg: %v", msg)
	}

	var dataLogs []*pb.Log

	for _, l := range logs {
		dataLogs = append(dataLogs, &pb.Log{
			Hash:     l.Hash.String,
			ItemId:   l.ItemID,
			ItemType: l.ItemType,
			Status:   pb.LaunchLogStatus(pb.LaunchLogStatus_value[l.Status]),
			GasPrice: l.GasPrice.String(),
		})
	}

	return &pb.GetReply{
		Status: pb.RequestStatus_REQUEST_SUCCESSFUL,
		Data:   dataLogs,
	}, nil
}

// the launcher has it's own watcher, no need to notify
func (*server) Notify(ctx context.Context, msg *pb.NotifyMessage) (*pb.NotifyReply, error) {
	return nil, fmt.Errorf("no implement")
}

func getSubscribeHubKey(itemType, itemId string) string {
	return fmt.Sprintf("Type:%s-ID:%s", itemType, itemId)
}

func sendLogStatusToSubscriber(log *LaunchLog, status pb.LaunchLogStatus) {
	key := getSubscribeHubKey(log.ItemType, log.ItemID)

	data, ok := subscribeHub.data[key]
	if !ok || data == nil {
		return
	}

	for s, _ := range data {
		_ = s.Send(&pb.SubscribeReply{
			Status:   status,
			Hash:     log.Hash.String,
			ItemId:   log.ItemID,
			ItemType: log.ItemType,
			ErrMsg:   log.ErrMsg,
		})
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
	data map[string]map[pb.Launcher_SubscribeServer]bool
}

func (sb *SubscribeHub) Register(key string, server pb.Launcher_SubscribeServer) {
	sb.m.Lock()
	defer sb.m.Unlock()

	if _, ok := sb.data[key]; !ok {
		sb.data[key] = make(map[pb.Launcher_SubscribeServer]bool)
	}

	sb.data[key][server] = true
}

func (sb *SubscribeHub) Remove(key string, server pb.Launcher_SubscribeServer) {
	sb.m.Lock()
	defer sb.m.Unlock()

	if _, ok := sb.data[key]; !ok {
		return
	}

	delete(sb.data[key], server)

	if len(sb.data[key]) == 0 {
		delete(sb.data, key)
	}
}

var subscribeHub *SubscribeHub

func startGrpcServer(ctx context.Context) {
	subscribeHub = &SubscribeHub{
		m:    &sync.Mutex{},
		data: make(map[string]map[pb.Launcher_SubscribeServer]bool),
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
