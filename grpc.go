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
	"strconv"
	"strings"
	"sync"
)

// notify the send loop to start
var newRequestChannel = make(chan int, 100)

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

		if !value.Equal(value.Round(0)) {
			return nil, fmt.Errorf("value must be an integer, not a decimal")
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
		Status:   pb.LaunchLogStatus_CREATED.String(),
	}

	if err = db.Create(log).Error; err != nil {
		return nil, err
	}

	key := getSubscribeHubKey(msg.ItemType, msg.ItemId)

	resCh := make(chan *pb.CreateReply, 1)
	errCh := make(chan error, 1)

	cb := func(l *LaunchLog, err error) {
		logrus.Infof("Create callback for log %d, error: %+v", l.ID, err)
		if err != nil {
			errCh <- err
			return
		}

		resCh <- &pb.CreateReply{
			Status: pb.RequestStatus_REQUEST_SUCCESSFUL,
			ErrMsg: "",
			Data: &pb.Log{
				Hash:     l.Hash.String,
				ItemId:   l.ItemID,
				ItemType: l.ItemType,
				Status:   pb.LaunchLogStatus(pb.LaunchLogStatus_value[l.Status]),
				GasPrice: l.GasPrice.String(),
				GasLimit: strconv.FormatUint(l.GasLimit, 10),
			},
		}
	}

	subscribeHub.Register(key, &cb)
	defer subscribeHub.Remove(key, &cb)

	// notify the send loop to work
	newRequestChannel <- 1

	select {
	case err := <-errCh:
		return nil, err
	case res := <-resCh:
		return res, nil
	}
}

type CreateCallbackFunc func(l *LaunchLog, err error)

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
			Hash:       l.Hash.String,
			ItemId:     l.ItemID,
			ItemType:   l.ItemType,
			Status:     pb.LaunchLogStatus(pb.LaunchLogStatus_value[l.Status]),
			GasPrice:   l.GasPrice.String(),
			GasUsed:    l.GasUsed,
			ExecutedAt: l.ExecutedAt,
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

func sendLogStatusToSubscriber(log *LaunchLog, err error) {
	logrus.Infof("sendLogStatusToSubscriber for log %d", log.ID)

	key := getSubscribeHubKey(log.ItemType, log.ItemID)

	data, ok := subscribeHub.data[key]

	if !ok || data == nil {
		return
	}

	for s, _ := range data {
		switch v := s.(type) {
		case pb.Launcher_SubscribeServer:
			_ = v.Send(&pb.SubscribeReply{
				Status:   pb.LaunchLogStatus(pb.LaunchLogStatus_value[log.Status]),
				Hash:     log.Hash.String,
				ItemId:   log.ItemID,
				ItemType: log.ItemType,
				ErrMsg:   log.ErrMsg,
			})
		case *CreateCallbackFunc:
			(*v)(log, err)
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

func startGrpcServer(ctx context.Context) {
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
