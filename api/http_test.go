package api

import (
	"context"
	"testing"
)

func TestStartHTTPServer(t *testing.T) {
	StartHTTPServer(context.Background())
}
