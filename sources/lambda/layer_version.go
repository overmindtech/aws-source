package lambda

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func layerVersionGetInputMapper(scope, query string) *lambda.GetLayerVersionInput {
	sections := strings.Split(query, ":")

	if len(sections) < 2 {
		return nil
	}

	version := sections[len(sections)-1]
	name := strings.Join(sections[0:len(sections)-1], ":")
	versionInt, err := strconv.Atoi(version)

	if err != nil {
		return nil
	}

	return &lambda.GetLayerVersionInput{
		LayerName:     &name,
		VersionNumber: *sources.PtrInt64(int64(versionInt)),
	}
}

func layerVersionGetFunc(ctx context.Context, client LambdaClient, scope string, input *lambda.GetLayerVersionInput) (*sdp.Item, error) {
	if input == nil {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOTFOUND,
			ErrorString: "nil input provided to query",
		}
	}

	out, err := client.GetLayerVersion(ctx, input)

	if err != nil {
		return nil, err
	}

	attributes, err := sources.ToAttributesCase(out, "resultMetadata")

	if err != nil {
		return nil, err
	}

	err = attributes.Set("fullName", fmt.Sprintf("%v:%v", *input.LayerName, input.VersionNumber))

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "lambda-layer-version",
		UniqueAttribute: "fullName",
		Attributes:      attributes,
		Scope:           scope,
	}

	var a *sources.ARN

	if out.Content != nil {
		if out.Content.SigningJobArn != nil {
			if a, err = sources.ParseARN(*out.Content.SigningJobArn); err == nil {
				// +overmind:link signer-signing-job
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "signer-signing-job",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *out.Content.SigningJobArn,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Signing jobs can affect layers
						In: true,
						// Changing the layer won't affect the signing job
						Out: false,
					},
				})
			}
		}

		if out.Content.SigningProfileVersionArn != nil {
			if a, err = sources.ParseARN(*out.Content.SigningProfileVersionArn); err == nil {
				// +overmind:link signer-signing-profile
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "signer-signing-profile",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *out.Content.SigningProfileVersionArn,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Signing profiles can affect layers
						In: true,
						// Changing the layer won't affect the signing profile
						Out: false,
					},
				})
			}
		}
	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type lambda-layer-version
// +overmind:descriptiveType Lambda Layer Version
// +overmind:get Get a layer version by full name ({layerName}:{versionNumber})
// +overmind:search Search for layer versions by ARN
// +overmind:group AWS

func NewLayerVersionSource(config aws.Config, accountID string, region string) *sources.AlwaysGetSource[*lambda.ListLayerVersionsInput, *lambda.ListLayerVersionsOutput, *lambda.GetLayerVersionInput, *lambda.GetLayerVersionOutput, LambdaClient, *lambda.Options] {
	return &sources.AlwaysGetSource[*lambda.ListLayerVersionsInput, *lambda.ListLayerVersionsOutput, *lambda.GetLayerVersionInput, *lambda.GetLayerVersionOutput, LambdaClient, *lambda.Options]{
		ItemType:       "lambda-layer-version",
		Client:         lambda.NewFromConfig(config),
		AccountID:      accountID,
		Region:         region,
		DisableList:    true,
		GetInputMapper: layerVersionGetInputMapper,
		GetFunc:        layerVersionGetFunc,
		ListInput:      &lambda.ListLayerVersionsInput{},
		ListFuncOutputMapper: func(output *lambda.ListLayerVersionsOutput, input *lambda.ListLayerVersionsInput) ([]*lambda.GetLayerVersionInput, error) {
			return []*lambda.GetLayerVersionInput{}, nil
		},
		ListFuncPaginatorBuilder: func(client LambdaClient, input *lambda.ListLayerVersionsInput) sources.Paginator[*lambda.ListLayerVersionsOutput, *lambda.Options] {
			return lambda.NewListLayerVersionsPaginator(client, input)
		},
	}
}
