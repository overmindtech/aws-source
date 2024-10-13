package ecs

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/overmindtech/aws-source/adapterhelpers"
)

type TestClient struct{}

func GetAutoConfig(t *testing.T) (*ecs.Client, string, string) {
	config, account, region := adapterhelpers.GetAutoConfig(t)
	client := ecs.NewFromConfig(config)

	return client, account, region
}
