package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func iamInstanceProfileAssociationOutputMapper(_ context.Context, _ *ec2.Client, scope string, _ *ec2.DescribeIamInstanceProfileAssociationsInput, output *ec2.DescribeIamInstanceProfileAssociationsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, assoc := range output.IamInstanceProfileAssociations {
		attributes, err := sources.ToAttributesCase(assoc)

		if err != nil {
			return nil, err
		}

		item := sdp.Item{
			Type:            "ec2-iam-instance-profile-association",
			UniqueAttribute: "associationId",
			Attributes:      attributes,
			Scope:           scope,
		}

		if assoc.IamInstanceProfile != nil && assoc.IamInstanceProfile.Arn != nil {
			if arn, err := sources.ParseARN(*assoc.IamInstanceProfile.Arn); err == nil {
				// +overmind:link iam-instance-profile
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "iam-instance-profile",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *assoc.IamInstanceProfile.Arn,
						Scope:  sources.FormatScope(arn.AccountID, arn.Region),
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changes to the profile will affect this
						In: true,
						// We can't affect the profile
						Out: false,
					},
				})
			}
		}

		if assoc.InstanceId != nil {
			// +overmind:link ec2-instance
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ec2-instance",
					Method: sdp.QueryMethod_GET,
					Query:  *assoc.InstanceId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changes to the instance will not affect the association
					In: false,
					// changes to the association will affect the instance
					Out: true,
				},
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type ec2-iam-instance-profile-association
// +overmind:descriptiveType IAM Instance Profile Association
// +overmind:get Get an IAM Instance Profile Association
// +overmind:list List IAM Instance Profile Associations
// +overmind:search Search IAM Instance Profile Associations by ARN
// +overmind:group AWS

// NewIamInstanceProfileAssociationSource Creates a new source for aws-IamInstanceProfileAssociation resources
func NewIamInstanceProfileAssociationSource(config aws.Config, accountID string, limit *sources.LimitBucket) *sources.DescribeOnlySource[*ec2.DescribeIamInstanceProfileAssociationsInput, *ec2.DescribeIamInstanceProfileAssociationsOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeIamInstanceProfileAssociationsInput, *ec2.DescribeIamInstanceProfileAssociationsOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		Client:    ec2.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "ec2-iam-instance-profile-association",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeIamInstanceProfileAssociationsInput) (*ec2.DescribeIamInstanceProfileAssociationsOutput, error) {
			limit.Wait(ctx) // Wait for rate limiting // Wait for late limiting
			return client.DescribeIamInstanceProfileAssociations(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*ec2.DescribeIamInstanceProfileAssociationsInput, error) {
			return &ec2.DescribeIamInstanceProfileAssociationsInput{
				AssociationIds: []string{query},
			}, nil
		},
		InputMapperList: func(scope string) (*ec2.DescribeIamInstanceProfileAssociationsInput, error) {
			return &ec2.DescribeIamInstanceProfileAssociationsInput{}, nil
		},
		OutputMapper: iamInstanceProfileAssociationOutputMapper,
	}
}
