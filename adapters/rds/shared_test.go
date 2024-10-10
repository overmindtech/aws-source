package rds

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/overmindtech/aws-source/adapters"
)

func GetAutoConfig(t *testing.T) (*rds.Client, string, string) {
	config, account, region := adapters.GetAutoConfig(t)
	client := rds.NewFromConfig(config)

	return client, account, region
}
