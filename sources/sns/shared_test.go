package sns

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/overmindtech/aws-source/sources"
)

func GetAutoConfig(t *testing.T) (*sns.Client, string, string) {
	config, account, region := sources.GetAutoConfig(t)
	client := sns.NewFromConfig(config)

	return client, account, region
}
