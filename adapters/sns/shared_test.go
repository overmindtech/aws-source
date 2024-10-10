package sns

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/overmindtech/aws-source/adapters"
)

func GetAutoConfig(t *testing.T) (*sns.Client, string, string) {
	config, account, region := adapters.GetAutoConfig(t)
	client := sns.NewFromConfig(config)

	return client, account, region
}
