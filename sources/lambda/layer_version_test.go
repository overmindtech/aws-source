package lambda

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestLayerVersionGetInputMapper(t *testing.T) {
	tests := []struct {
		Query     string
		ExpectNil bool
	}{
		{
			Query:     "foo:1",
			ExpectNil: false,
		},
		{
			Query:     "foo:1:2",
			ExpectNil: false,
		},
		{
			Query:     "",
			ExpectNil: true,
		},
		{
			Query:     "bar",
			ExpectNil: true,
		},
		{
			Query:     ":",
			ExpectNil: true,
		},
	}

	for _, test := range tests {
		t.Run(test.Query, func(t *testing.T) {
			input := LayerVersionGetInputMapper("foo", test.Query)

			if input == nil && !test.ExpectNil {
				t.Error("input was nil unexpectedly")
			}

			if input != nil && test.ExpectNil {
				t.Error("input was non-nil when expected to be nil")
			}
		})
	}
}

func (t *TestLambdaClient) GetLayerVersion(ctx context.Context, params *lambda.GetLayerVersionInput, optFns ...func(*lambda.Options)) (*lambda.GetLayerVersionOutput, error) {
	return &lambda.GetLayerVersionOutput{
		CompatibleArchitectures: []types.Architecture{
			types.ArchitectureArm64,
		},
		CompatibleRuntimes: []types.Runtime{
			types.RuntimeDotnet6,
		},
		Content: &types.LayerVersionContentOutput{
			CodeSha256:               sources.PtrString("sha"),
			CodeSize:                 100,
			Location:                 sources.PtrString("somewhere"),
			SigningJobArn:            sources.PtrString("arn:aws:service:region:account:type/id"),
			SigningProfileVersionArn: sources.PtrString("arn:aws:service:region:account:type/id"),
		},
		CreatedDate:     sources.PtrString("YYYY-MM-DDThh:mm:ss.sTZD"),
		Description:     sources.PtrString("description"),
		LayerArn:        sources.PtrString("arn:aws:service:region:account:type/id"),
		LayerVersionArn: sources.PtrString("arn:aws:service:region:account:type/id"),
		LicenseInfo:     sources.PtrString("info"),
		Version:         params.VersionNumber,
	}, nil
}

func (t *TestLambdaClient) ListLayerVersions(context.Context, *lambda.ListLayerVersionsInput, ...func(*lambda.Options)) (*lambda.ListLayerVersionsOutput, error) {
	return &lambda.ListLayerVersionsOutput{}, nil
}

func TestLayerVersionGetFunc(t *testing.T) {
	item, err := LayerVersionGetFunc(context.Background(), &TestLambdaClient{}, "foo", &lambda.GetLayerVersionInput{
		LayerName:     sources.PtrString("layer"),
		VersionNumber: 999,
	})

	if err != nil {
		t.Error(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

	tests := sources.ItemRequestTests{
		{
			ExpectedType:   "signer-signing-job",
			ExpectedMethod: sdp.RequestMethod_SEARCH,
			ExpectedQuery:  "arn:aws:service:region:account:type/id",
			ExpectedScope:  "account.region",
		},
		{
			ExpectedType:   "signer-signing-profile",
			ExpectedMethod: sdp.RequestMethod_SEARCH,
			ExpectedQuery:  "arn:aws:service:region:account:type/id",
			ExpectedScope:  "account.region",
		},
	}

	tests.Execute(t, item)
}

func TestNewLayerVersionSource(t *testing.T) {
	config, account, region := sources.GetAutoConfig(t)

	source := NewLayerVersionSource(config, account, region)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
