package gas

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestFetcher(t *testing.T){
	go StartFetcher(context.Background())
	<-time.After(time.Second * 5)

	price,err := Get()
	assert.Nil(t, err)
	spew.Dump(price)
}
