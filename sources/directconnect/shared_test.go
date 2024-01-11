package directconnect

import (
	"context"
	"os"
	"testing"

	"github.com/overmindtech/aws-source/sources"
)

var TestRateLimit = sources.LimitBucket{
	MaxCapacity: 50,
	RefillRate:  20,
}

func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	TestRateLimit.Start(ctx)
	os.Exit(m.Run())
}
