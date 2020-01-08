package api

import (
	"database/sql"
	"fmt"
	"git.ddex.io/infrastructure/ethereum-launcher/gas"
	pb "git.ddex.io/infrastructure/ethereum-launcher/messages"
	"git.ddex.io/infrastructure/ethereum-launcher/models"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

// notify the send loop to start
var NewRequestChannel = make(chan int, 100)

func createLog(msg *pb.CreateMessage) (*pb.CreateReply, error) {
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
		normalPrice,urgentPrice := gas.GetCurrentGasPrice()
		if msg.IsUrgent {
			gasPrice = urgentPrice
		} else {
			gasPrice = normalPrice
		}
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
	if err := models.DB.Model(&models.LaunchLog{}).Where("item_type = ? and item_id = ?", msg.ItemType, msg.ItemId).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("get item_type and item_id count error %v", err)
	}

	if count > 0 {
		return nil, fmt.Errorf("item_type and item_id exists !!")
	}

	log := &models.LaunchLog{
		Hash: sql.NullString{
			Valid: false,
		},
		IsUrgent: msg.IsUrgent,
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

	if err = models.DB.Create(log).Error; err != nil {
		return nil, err
	}

	key := getSubscribeHubKey(msg.ItemType, msg.ItemId)

	resCh := make(chan *pb.CreateReply, 1)
	errCh := make(chan error, 1)

	// it's important to use var here
	// otherwise, golang cant's cast the pointer back
	var cb CreateCallbackFunc

	cb = func(l *models.LaunchLog, err error) {
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
				Error:    l.ErrMsg,
			},
		}
	}

	subscribeHub.Register(key, &cb)
	defer subscribeHub.Remove(key, &cb)

	// notify the send loop to work
	NewRequestChannel <- 1

	select {
	case err := <-errCh:
		return nil, err
	case res := <-resCh:
		return res, nil
	}
}

func getLog(msg *pb.GetMessage) (*pb.GetReply, error) {
	var logs []*models.LaunchLog

	if msg.Hash != "" {
		models.DB.Where("hash = ?", msg.Hash).Find(&logs)
	} else if msg.ItemType != "" && msg.ItemId != "" {
		models.DB.Where("item_type = ? and item_id = ?", msg.ItemType, msg.ItemId).Find(&logs)
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
			Error:      l.ErrMsg,
		})
	}

	return &pb.GetReply{
		Status: pb.RequestStatus_REQUEST_SUCCESSFUL,
		Data:   dataLogs,
	}, nil
}
