package iam

import (
	"context"
	"os"
	"testing"

	"github.com/overmindtech/aws-source/sources"
)

// TestIAMClient Test client that returns three pages
type TestIAMClient struct{}

var TestRateLimit = sources.LimitBucket{
	MaxCapacity: 20,
	RefillRate:  15,
}

func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	TestRateLimit.Start(ctx)
	os.Exit(m.Run())
}
