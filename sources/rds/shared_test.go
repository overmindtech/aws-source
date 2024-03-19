package rds

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/overmindtech/aws-source/sources"
)

func GetAutoConfig(t *testing.T) (*rds.Client, string, string) {
	config, account, region := sources.GetAutoConfig(t)
	client := rds.NewFromConfig(config)

	return client, account, region
}
