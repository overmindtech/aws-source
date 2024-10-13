package lambda

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/overmindtech/aws-source/adapterhelpers"
)

type TestLambdaClient struct{}

func GetAutoConfig(t *testing.T) (*lambda.Client, string, string) {
	config, account, region := adapterhelpers.GetAutoConfig(t)
	client := lambda.NewFromConfig(config)

	return client, account, region
}
