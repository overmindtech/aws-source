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

func LayerVersionGetInputMapper(scope, query string) *lambda.GetLayerVersionInput {
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

func LayerVersionGetFunc(ctx context.Context, client LambdaClient, scope string, input *lambda.GetLayerVersionInput) (*sdp.Item, error) {
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
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "signer-signing-job",
					Method: sdp.RequestMethod_SEARCH,
					Query:  *out.Content.SigningJobArn,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				})
			}
		}

		if out.Content.SigningProfileVersionArn != nil {
			if a, err = sources.ParseARN(*out.Content.SigningProfileVersionArn); err == nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "signer-signing-profile",
					Method: sdp.RequestMethod_SEARCH,
					Query:  *out.Content.SigningProfileVersionArn,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				})
			}
		}
	}

	return &item, nil
}

func NewLayerVersionSource(config aws.Config, accountID string, region string) *sources.AlwaysGetSource[*lambda.ListLayerVersionsInput, *lambda.ListLayerVersionsOutput, *lambda.GetLayerVersionInput, *lambda.GetLayerVersionOutput, LambdaClient, *lambda.Options] {
	return &sources.AlwaysGetSource[*lambda.ListLayerVersionsInput, *lambda.ListLayerVersionsOutput, *lambda.GetLayerVersionInput, *lambda.GetLayerVersionOutput, LambdaClient, *lambda.Options]{
		ItemType:       "lambda-layer-version",
		Client:         lambda.NewFromConfig(config),
		AccountID:      accountID,
		Region:         region,
		DisableList:    true,
		GetInputMapper: LayerVersionGetInputMapper,
		GetFunc:        LayerVersionGetFunc,
		ListInput:      &lambda.ListLayerVersionsInput{},
		ListFuncOutputMapper: func(output *lambda.ListLayerVersionsOutput, input *lambda.ListLayerVersionsInput) ([]*lambda.GetLayerVersionInput, error) {
			return []*lambda.GetLayerVersionInput{}, nil
		},
		ListFuncPaginatorBuilder: func(client LambdaClient, input *lambda.ListLayerVersionsInput) sources.Paginator[*lambda.ListLayerVersionsOutput, *lambda.Options] {
			return lambda.NewListLayerVersionsPaginator(client, input)
		},
	}
}
