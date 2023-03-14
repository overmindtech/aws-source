package autoscaling

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/aws-source/sources/ec2"
	"github.com/overmindtech/sdp-go"
)

func autoScalingGroupOutputMapper(scope string, _ *autoscaling.DescribeAutoScalingGroupsInput, output *autoscaling.DescribeAutoScalingGroupsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	var item sdp.Item
	var attributes *sdp.ItemAttributes
	var err error

	for _, asg := range output.AutoScalingGroups {
		attributes, err = sources.ToAttributesCase(asg)

		if err != nil {
			return nil, err
		}

		item = sdp.Item{
			Type:            "autoscaling-auto-scaling-group",
			UniqueAttribute: "autoScalingGroupName",
			Scope:           scope,
			Attributes:      attributes,
		}

		if asg.MixedInstancesPolicy != nil {
			if asg.MixedInstancesPolicy.LaunchTemplate != nil {
				if asg.MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification != nil {
					if asg.MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification.LaunchTemplateId != nil {
						item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
							Type:   "ec2-launch-template",
							Method: sdp.QueryMethod_GET,
							Query:  *asg.MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification.LaunchTemplateId,
							Scope:  scope,
						})
					}
				}
			}
		}

		var a *sources.ARN
		var err error

		for _, tgARN := range asg.TargetGroupARNs {
			if a, err = sources.ParseARN(tgARN); err == nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "elbv2-target-group",
					Method: sdp.QueryMethod_SEARCH,
					Query:  tgARN,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				})
			}
		}

		for _, instance := range asg.Instances {
			if instance.InstanceId != nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "ec2-instance",
					Method: sdp.QueryMethod_GET,
					Query:  *instance.InstanceId,
					Scope:  scope,
				})
			}

			if instance.LaunchTemplate != nil {
				if instance.LaunchTemplate.LaunchTemplateId != nil {
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
						Type:   "ec2-launch-template",
						Method: sdp.QueryMethod_GET,
						Query:  *instance.LaunchTemplate.LaunchTemplateId,
						Scope:  scope,
					})
				}
			}
		}

		if asg.ServiceLinkedRoleARN != nil {
			if a, err = sources.ParseARN(*asg.ServiceLinkedRoleARN); err == nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "iam-role",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *asg.ServiceLinkedRoleARN,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				})
			}
		}

		if asg.LaunchConfigurationName != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "autoscaling-launch-configuration",
				Method: sdp.QueryMethod_GET,
				Query:  *asg.LaunchConfigurationName,
				Scope:  scope,
			})
		}

		if asg.LaunchTemplate != nil {
			if asg.LaunchTemplate.LaunchTemplateId != nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "ec2-launch-template",
					Method: sdp.QueryMethod_GET,
					Query:  *asg.LaunchTemplate.LaunchTemplateId,
					Scope:  scope,
				})
			}
		}

		if asg.PlacementGroup != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "ec2-placement-group",
				Method: sdp.QueryMethod_GET,
				Query:  *asg.PlacementGroup,
				Scope:  scope,
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

func NewAutoScalingGroupSource(config aws.Config, accountID string, limit *ec2.LimitBucket) *sources.DescribeOnlySource[*autoscaling.DescribeAutoScalingGroupsInput, *autoscaling.DescribeAutoScalingGroupsOutput, *autoscaling.Client, *autoscaling.Options] {
	return &sources.DescribeOnlySource[*autoscaling.DescribeAutoScalingGroupsInput, *autoscaling.DescribeAutoScalingGroupsOutput, *autoscaling.Client, *autoscaling.Options]{
		ItemType:  "autoscaling-auto-scaling-group",
		Config:    config,
		AccountID: accountID,
		Client:    autoscaling.NewFromConfig(config),
		InputMapperGet: func(scope, query string) (*autoscaling.DescribeAutoScalingGroupsInput, error) {
			return &autoscaling.DescribeAutoScalingGroupsInput{
				AutoScalingGroupNames: []string{query},
			}, nil
		},
		InputMapperList: func(scope string) (*autoscaling.DescribeAutoScalingGroupsInput, error) {
			return &autoscaling.DescribeAutoScalingGroupsInput{}, nil
		},
		PaginatorBuilder: func(client *autoscaling.Client, params *autoscaling.DescribeAutoScalingGroupsInput) sources.Paginator[*autoscaling.DescribeAutoScalingGroupsOutput, *autoscaling.Options] {
			return autoscaling.NewDescribeAutoScalingGroupsPaginator(client, params)
		},
		DescribeFunc: func(ctx context.Context, client *autoscaling.Client, input *autoscaling.DescribeAutoScalingGroupsInput) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
			return client.DescribeAutoScalingGroups(ctx, input)
		},
		OutputMapper: autoScalingGroupOutputMapper,
	}
}
