package lambda

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

type FunctionDetails struct {
	Code               *types.FunctionCodeLocation
	Concurrency        *types.Concurrency
	Configuration      *types.FunctionConfiguration
	UrlConfigs         []*types.FunctionUrlConfig
	EventInvokeConfigs []*types.FunctionEventInvokeConfig
	Tags               map[string]string
}

// FunctionGetFunc Gets the details of a specific lambda function
func functionGetFunc(ctx context.Context, client LambdaClient, scope string, input *lambda.GetFunctionInput) (*sdp.Item, error) {
	out, err := client.GetFunction(ctx, input)

	if err != nil {
		return nil, err
	}

	if out.Configuration == nil {
		return nil, errors.New("function has nil configuration")
	}

	if out.Configuration.FunctionName == nil {
		return nil, errors.New("function has empty name")
	}

	function := FunctionDetails{
		Code:          out.Code,
		Concurrency:   out.Concurrency,
		Configuration: out.Configuration,
		Tags:          out.Tags,
	}

	// Get details of all URL configs
	urlConfigs := lambda.NewListFunctionUrlConfigsPaginator(client, &lambda.ListFunctionUrlConfigsInput{
		FunctionName: out.Configuration.FunctionName,
	})

	var urlOut *lambda.ListFunctionUrlConfigsOutput

	for urlConfigs.HasMorePages() {
		urlOut, err = urlConfigs.NextPage(ctx)

		if err != nil {
			continue
		}

		for _, config := range urlOut.FunctionUrlConfigs {
			function.UrlConfigs = append(function.UrlConfigs, &config)
		}
	}

	// Get details of event configs
	eventConfigs := lambda.NewListFunctionEventInvokeConfigsPaginator(client, &lambda.ListFunctionEventInvokeConfigsInput{
		FunctionName: out.Configuration.FunctionName,
	})

	var eventOut *lambda.ListFunctionEventInvokeConfigsOutput

	for eventConfigs.HasMorePages() {
		eventOut, err = eventConfigs.NextPage(ctx)

		if err != nil {
			continue
		}

		for _, event := range eventOut.FunctionEventInvokeConfigs {
			function.EventInvokeConfigs = append(function.EventInvokeConfigs, &event)
		}
	}

	attributes, err := sources.ToAttributesCase(function, "resultMetadata")

	if err != nil {
		return nil, err
	}

	err = attributes.Set("name", *out.Configuration.FunctionName)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "lambda-function",
		UniqueAttribute: "name",
		Attributes:      attributes,
		Scope:           scope,
	}

	if function.Code != nil {
		if function.Code.Location != nil {
			// +overmind:link http
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{Query: &sdp.Query{
				Type:   "http",
				Method: sdp.QueryMethod_GET,
				Query:  *function.Code.Location,
				Scope:  "global",
			}})
		}

		if function.Code.ImageUri != nil {
			// +overmind:link http
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{Query: &sdp.Query{
				Type:   "http",
				Method: sdp.QueryMethod_GET,
				Query:  *function.Code.ImageUri,
				Scope:  "global",
			}})
		}

		if function.Code.ResolvedImageUri != nil {
			// +overmind:link http
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{Query: &sdp.Query{
				Type:   "http",
				Method: sdp.QueryMethod_GET,
				Query:  *function.Code.ResolvedImageUri,
				Scope:  "global",
			}})
		}
	}

	var a *sources.ARN

	if function.Configuration != nil {
		switch function.Configuration.State {
		case types.StatePending:
			item.Health = sdp.Health_HEALTH_PENDING.Enum()
		case types.StateActive:
			item.Health = sdp.Health_HEALTH_OK.Enum()
		case types.StateInactive:
			item.Health = nil
		case types.StateFailed:
			item.Health = sdp.Health_HEALTH_ERROR.Enum()
		}

		if function.Configuration.Role != nil {
			if a, err = sources.ParseARN(*function.Configuration.Role); err == nil {
				// +overmind:link iam-role
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{Query: &sdp.Query{
					Type:   "iam-role",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *function.Configuration.Role,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				}})
			}
		}

		if function.Configuration.DeadLetterConfig != nil {
			if function.Configuration.DeadLetterConfig.TargetArn != nil {
				if req, err := GetEventLinkedItem(*function.Configuration.DeadLetterConfig.TargetArn); err == nil {
					item.LinkedItemQueries = append(item.LinkedItemQueries, req)
				}
			}
		}

		for _, fsConfig := range function.Configuration.FileSystemConfigs {
			if fsConfig.Arn != nil {
				if a, err = sources.ParseARN(*fsConfig.Arn); err == nil {
					// +overmind:link efs-access-point
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{Query: &sdp.Query{
						Type:   "efs-access-point",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *fsConfig.Arn,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					}})
				}
			}
		}

		if function.Configuration.KMSKeyArn != nil {
			if a, err = sources.ParseARN(*function.Configuration.KMSKeyArn); err == nil {
				// +overmind:link kms-key
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{Query: &sdp.Query{
					Type:   "kms-key",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *function.Configuration.KMSKeyArn,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				}})
			}
		}

		for _, layer := range function.Configuration.Layers {
			if layer.Arn != nil {
				if a, err = sources.ParseARN(*layer.Arn); err == nil {
					// Strip the leading "layer:"
					name := strings.TrimPrefix(a.Resource, "layer:")

					// +overmind:link lambda-layer-version
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{Query: &sdp.Query{
						Type:   "lambda-layer-version",
						Method: sdp.QueryMethod_GET,
						Query:  name,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					}})
				}
			}

			if layer.SigningJobArn != nil {
				if a, err = sources.ParseARN(*layer.SigningJobArn); err == nil {
					// +overmind:link signer-signing-job
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{Query: &sdp.Query{
						Type:   "signer-signing-job",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *layer.SigningJobArn,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					}})
				}
			}

			if layer.SigningProfileVersionArn != nil {
				if a, err = sources.ParseARN(*layer.SigningProfileVersionArn); err == nil {
					// +overmind:link signer-signing-profile
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{Query: &sdp.Query{
						Type:   "signer-signing-profile",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *layer.SigningProfileVersionArn,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					}})
				}
			}
		}

		if function.Configuration.MasterArn != nil {
			if a, err = sources.ParseARN(*function.Configuration.MasterArn); err == nil {
				// +overmind:link lambda-function
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{Query: &sdp.Query{
					Type:   "lambda-function",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *function.Configuration.MasterArn,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				}})
			}
		}

		if function.Configuration.SigningJobArn != nil {
			if a, err = sources.ParseARN(*function.Configuration.SigningJobArn); err == nil {
				// +overmind:link signer-signing-job
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{Query: &sdp.Query{
					Type:   "signer-signing-job",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *function.Configuration.SigningJobArn,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				}})
			}
		}

		if function.Configuration.SigningProfileVersionArn != nil {
			if a, err = sources.ParseARN(*function.Configuration.SigningProfileVersionArn); err == nil {
				// +overmind:link signer-signing-profile
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{Query: &sdp.Query{
					Type:   "signer-signing-profile",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *function.Configuration.SigningProfileVersionArn,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				}})
			}
		}

		if function.Configuration.VpcConfig != nil {
			for _, id := range function.Configuration.VpcConfig.SecurityGroupIds {
				// +overmind:link ec2-security-group
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{Query: &sdp.Query{
					Type:   "ec2-security-group",
					Method: sdp.QueryMethod_GET,
					Query:  id,
					Scope:  scope,
				}})
			}

			for _, id := range function.Configuration.VpcConfig.SubnetIds {
				// +overmind:link ec2-subnet
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{Query: &sdp.Query{
					Type:   "ec2-subnet",
					Method: sdp.QueryMethod_GET,
					Query:  id,
					Scope:  scope,
				}})
			}

			if function.Configuration.VpcConfig.VpcId != nil {
				// +overmind:link ec2-vpc
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{Query: &sdp.Query{
					Type:   "ec2-vpc",
					Method: sdp.QueryMethod_GET,
					Query:  *function.Configuration.VpcConfig.VpcId,
					Scope:  scope,
				}})
			}
		}
	}

	for _, config := range function.UrlConfigs {
		if config.FunctionUrl != nil {
			// +overmind:link http
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{Query: &sdp.Query{
				Type:   "http",
				Method: sdp.QueryMethod_GET,
				Query:  *config.FunctionUrl,
				Scope:  "global",
			}})
		}
	}

	for _, config := range function.EventInvokeConfigs {
		if config.DestinationConfig != nil {
			if config.DestinationConfig.OnFailure != nil {
				if config.DestinationConfig.OnFailure.Destination != nil {
					// Possible links from `GetEventLinkedItem()`
					// +overmind:link sns-topic
					// +overmind:link sqs-queue
					// +overmind:link lambda-function
					// +overmind:link events-event-bus

					lir, err := GetEventLinkedItem(*config.DestinationConfig.OnFailure.Destination)

					if err == nil {
						item.LinkedItemQueries = append(item.LinkedItemQueries, lir)
					}
				}
			}

			if config.DestinationConfig.OnSuccess != nil {
				if config.DestinationConfig.OnSuccess.Destination != nil {
					lir, err := GetEventLinkedItem(*config.DestinationConfig.OnSuccess.Destination)

					if err == nil {
						item.LinkedItemQueries = append(item.LinkedItemQueries, lir)
					}

				}
			}
		}
	}

	return &item, nil
}

// GetEventLinkedItem Gets the linked item request for a given destination ARN
func GetEventLinkedItem(destinationARN string) (*sdp.LinkedItemQuery, error) {
	parsed, err := sources.ParseARN(destinationARN)

	if err != nil {
		return nil, err
	}

	scope := sources.FormatScope(parsed.AccountID, parsed.Region)

	switch parsed.Service {
	case "sns":
		// In this case it's an SNS topic
		return &sdp.LinkedItemQuery{Query: &sdp.Query{
			Type:   "sns-topic",
			Method: sdp.QueryMethod_SEARCH,
			Query:  destinationARN,
			Scope:  scope,
		}}, nil
	case "sqs":
		return &sdp.LinkedItemQuery{Query: &sdp.Query{
			Type:   "sqs-queue",
			Method: sdp.QueryMethod_SEARCH,
			Query:  destinationARN,
			Scope:  scope,
		}}, nil
	case "lambda":
		return &sdp.LinkedItemQuery{Query: &sdp.Query{
			Type:   "lambda-function",
			Method: sdp.QueryMethod_SEARCH,
			Query:  destinationARN,
			Scope:  scope,
		}}, nil
	case "events":
		return &sdp.LinkedItemQuery{Query: &sdp.Query{
			Type:   "events-event-bus",
			Method: sdp.QueryMethod_SEARCH,
			Query:  destinationARN,
			Scope:  scope,
		}}, nil
	}

	return nil, errors.New("could not find matching request")
}

//go:generate docgen ../../docs-data
// +overmind:type lambda-function
// +overmind:descriptiveType Lambda Function
// +overmind:get Get a lambda function by name
// +overmind:list List all lambda functions
// +overmind:search Search for lambda functions by ARN
// +overmind:group AWS

func NewFunctionSource(config aws.Config, accountID string, region string) *sources.AlwaysGetSource[*lambda.ListFunctionsInput, *lambda.ListFunctionsOutput, *lambda.GetFunctionInput, *lambda.GetFunctionOutput, LambdaClient, *lambda.Options] {
	return &sources.AlwaysGetSource[*lambda.ListFunctionsInput, *lambda.ListFunctionsOutput, *lambda.GetFunctionInput, *lambda.GetFunctionOutput, LambdaClient, *lambda.Options]{
		ItemType:  "lambda-function",
		Client:    lambda.NewFromConfig(config),
		AccountID: accountID,
		Region:    region,
		ListInput: &lambda.ListFunctionsInput{},
		GetFunc:   functionGetFunc,
		GetInputMapper: func(scope, query string) *lambda.GetFunctionInput {
			return &lambda.GetFunctionInput{
				FunctionName: &query,
			}
		},
		ListFuncPaginatorBuilder: func(client LambdaClient, input *lambda.ListFunctionsInput) sources.Paginator[*lambda.ListFunctionsOutput, *lambda.Options] {
			return lambda.NewListFunctionsPaginator(client, input)
		},
		ListFuncOutputMapper: func(output *lambda.ListFunctionsOutput, input *lambda.ListFunctionsInput) ([]*lambda.GetFunctionInput, error) {
			inputs := make([]*lambda.GetFunctionInput, len(output.Functions))

			for i := range output.Functions {
				inputs[i] = &lambda.GetFunctionInput{
					FunctionName: output.Functions[i].FunctionName,
				}
			}

			return inputs, nil
		},
	}
}
