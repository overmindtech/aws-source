package directconnect

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	"github.com/overmindtech/aws-source/sources"
)

func GetAutoConfig(t *testing.T) (*directconnect.Client, string, string) {
	config, account, region := sources.GetAutoConfig(t)
	client := directconnect.NewFromConfig(config)

	return client, account, region
}
