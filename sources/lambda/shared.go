package lambda

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

// LambdaClient Represents the client we need to talk to Lambda, usually this is
// *lambda.Client
type LambdaClient interface {
	GetFunction(ctx context.Context, params *lambda.GetFunctionInput, optFns ...func(*lambda.Options)) (*lambda.GetFunctionOutput, error)
	GetLayerVersion(ctx context.Context, params *lambda.GetLayerVersionInput, optFns ...func(*lambda.Options)) (*lambda.GetLayerVersionOutput, error)

	lambda.ListFunctionEventInvokeConfigsAPIClient
	lambda.ListFunctionUrlConfigsAPIClient
	lambda.ListFunctionsAPIClient
	lambda.ListLayerVersionsAPIClient
}
