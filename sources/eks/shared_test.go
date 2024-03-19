package eks

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/overmindtech/aws-source/sources"
)

func GetAutoConfig(t *testing.T) (*eks.Client, string, string) {
	config, account, region := sources.GetAutoConfig(t)
	client := eks.NewFromConfig(config)

	return client, account, region
}
