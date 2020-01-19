package models

import (
	"testing"
)

func TestHandleLaunchLogStatus(t *testing.T) {
	// docker-compose -f docker-compose.yaml down -v
	// docker-compose -f docker-compose.yaml up db ethereum-node
	_ = ConnectDB(LocalDBUrl)
}

