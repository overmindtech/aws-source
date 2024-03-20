package route53

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/overmindtech/aws-source/sources"
)

func GetAutoConfig(t *testing.T) (*route53.Client, string, string) {
	config, account, region := sources.GetAutoConfig(t)
	client := route53.NewFromConfig(config)

	return client, account, region
}
