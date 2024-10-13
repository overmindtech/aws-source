package ec2

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/adapterhelpers"
)

func GetAutoConfig(t *testing.T) (*ec2.Client, string, string) {
	config, account, region := adapterhelpers.GetAutoConfig(t)
	client := ec2.NewFromConfig(config)

	return client, account, region
}
