package efs

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/efs"
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

func GetAutoConfig(t *testing.T) (*efs.Client, string, string) {
	config, account, region := sources.GetAutoConfig(t)
	client := efs.NewFromConfig(config)

	return client, account, region
}
