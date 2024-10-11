package efs

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/efs"
	"github.com/overmindtech/aws-source/adapters"
)

func GetAutoConfig(t *testing.T) (*efs.Client, string, string) {
	config, account, region := adapters.GetAutoConfig(t)
	client := efs.NewFromConfig(config)

	return client, account, region
}
