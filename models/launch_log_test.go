package models

import (
	"testing"
)

func TestHandleLaunchLogStatus(t *testing.T) {
	// docker-compose -f docker-compose.yaml up db
	_ = ConnectDB(LocalDBUrl)

}

