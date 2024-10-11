package directconnect

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	"github.com/overmindtech/aws-source/adapters"
)

func GetAutoConfig(t *testing.T) (*directconnect.Client, string, string) {
	config, account, region := adapters.GetAutoConfig(t)
	client := directconnect.NewFromConfig(config)

	return client, account, region
}
