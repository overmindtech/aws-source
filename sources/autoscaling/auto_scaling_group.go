package autoscaling

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func autoScalingGroupOutputMapper(_ context.Context, _ *autoscaling.Client, scope string, _ *autoscaling.DescribeAutoScalingGroupsInput, output *autoscaling.DescribeAutoScalingGroupsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	var item sdp.Item
	var attributes *sdp.ItemAttributes
	var err error

	for _, asg := range output.AutoScalingGroups {
		attributes, err = sources.ToAttributesWithExclude(asg)

		if err != nil {
			return nil, err
		}

		item = sdp.Item{
			Type:            "autoscaling-auto-scaling-group",
			UniqueAttribute: "AutoScalingGroupName",
			Scope:           scope,
			Attributes:      attributes,
		}

		tags := make(map[string]string)

		for _, tag := range asg.Tags {
			if tag.Key != nil && tag.Value != nil {
				tags[*tag.Key] = *tag.Value
			}
		}

		item.Tags = tags

		if asg.MixedInstancesPolicy != nil {
			if asg.MixedInstancesPolicy.LaunchTemplate != nil {
				if asg.MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification != nil {
					if asg.MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification.LaunchTemplateId != nil {
						// +overmind:link ec2-launch-template
						item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
							Query: &sdp.Query{
								Type:   "ec2-launch-template",
								Method: sdp.QueryMethod_GET,
								Query:  *asg.MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification.LaunchTemplateId,
								Scope:  scope,
							},
							BlastPropagation: &sdp.BlastPropagation{
								// Changes to a launch template will affect the ASG
								In: true,
								// Changes to an ASG won't affect the template
								Out: false,
							},
						})
					}
				}
			}
		}

		var a *sources.ARN
		var err error

		for _, tgARN := range asg.TargetGroupARNs {
			if a, err = sources.ParseARN(tgARN); err == nil {
				// +overmind:link elbv2-target-group
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "elbv2-target-group",
						Method: sdp.QueryMethod_SEARCH,
						Query:  tgARN,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changes to a target group won't affect the ASG
						In: false,
						// Changes to an ASG will affect the target group
						Out: true,
					},
				})
			}
		}

		for _, instance := range asg.Instances {
			if instance.InstanceId != nil {
				// +overmind:link ec2-instance
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ec2-instance",
						Method: sdp.QueryMethod_GET,
						Query:  *instance.InstanceId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changes to an instance could affect the ASG since it
						// could cause it to scale
						In: true,
						// Changes to an ASG can definitely affect an instance
						// since it might be terminated
						Out: true,
					},
				})
			}

			if instance.LaunchTemplate != nil {
				if instance.LaunchTemplate.LaunchTemplateId != nil {
					// +overmind:link ec2-launch-template
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "ec2-launch-template",
							Method: sdp.QueryMethod_GET,
							Query:  *instance.LaunchTemplate.LaunchTemplateId,
							Scope:  scope,
						},
						BlastPropagation: &sdp.BlastPropagation{
							// Changes to a launch template will affect the ASG
							In: true,
							// Changes to an ASG won't affect the template
							Out: false,
						},
					})
				}
			}
		}

		if asg.ServiceLinkedRoleARN != nil {
			if a, err = sources.ParseARN(*asg.ServiceLinkedRoleARN); err == nil {
				// +overmind:link iam-role
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "iam-role",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *asg.ServiceLinkedRoleARN,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changes to a role can affect the functioning of the
						// ASG
						In: true,
						// ASG changes wont affect the role though
						Out: false,
					},
				})
			}
		}

		if asg.LaunchConfigurationName != nil {
			// +overmind:link autoscaling-launch-configuration
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "autoscaling-launch-configuration",
					Method: sdp.QueryMethod_GET,
					Query:  *asg.LaunchConfigurationName,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Very tightly coupled
					In:  true,
					Out: true,
				},
			})
		}

		if asg.LaunchTemplate != nil {
			if asg.LaunchTemplate.LaunchTemplateId != nil {
				// +overmind:link ec2-launch-template
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ec2-launch-template",
						Method: sdp.QueryMethod_GET,
						Query:  *asg.LaunchTemplate.LaunchTemplateId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changes to a launch template will affect the ASG
						In: true,
						// Changes to an ASG won't affect the template
						Out: false,
					},
				})
			}
		}

		if asg.PlacementGroup != nil {
			// +overmind:link ec2-placement-group
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ec2-placement-group",
					Method: sdp.QueryMethod_GET,
					Query:  *asg.PlacementGroup,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changes to a placement group can affect the ASG
					In: true,
					// Changes to an ASG can affect the placement group
					Out: true,
				},
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

// +overmind:type autoscaling-auto-scaling-group
// +overmind:descriptiveType Autoscaling Group
// +overmind:get Get an Autoscaling Group by name
// +overmind:list List Autoscaling Groups
// +overmind:search Search for Autoscaling Groups by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_autoscaling_group.arn
// +overmind:terraform:method SEARCH
//
//go:generate docgen ../../docs-data
func NewAutoScalingGroupSource(client *autoscaling.Client, accountID string, region string) *sources.DescribeOnlySource[*autoscaling.DescribeAutoScalingGroupsInput, *autoscaling.DescribeAutoScalingGroupsOutput, *autoscaling.Client, *autoscaling.Options] {
	return &sources.DescribeOnlySource[*autoscaling.DescribeAutoScalingGroupsInput, *autoscaling.DescribeAutoScalingGroupsOutput, *autoscaling.Client, *autoscaling.Options]{
		ItemType:        "autoscaling-auto-scaling-group",
		AccountID:       accountID,
		Region:          region,
		Client:          client,
		AdapterMetadata: AutoScalingGroupMetadata(),
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

func AutoScalingGroupMetadata() sdp.AdapterMetadata {
	return sdp.AdapterMetadata{
		Type:            "autoscaling-auto-scaling-group",
		DescriptiveName: "Autoscaling Group",
		Category:        sdp.AdapterCategory_ADAPTER_CATEGORY_CONFIGURATION,
		SupportedQueryMethods: &sdp.AdapterSupportedQueryMethods{
			Get:               true,
			List:              true,
			Search:            true,
			GetDescription:    "Get an Autoscaling Group by name",
			ListDescription:   "List Autoscaling Groups",
			SearchDescription: "Search for Autoscaling Groups by ARN",
		},
		TerraformMappings: []*sdp.TerraformMapping{
			{
				TerraformQueryMap: "aws_autoscaling_group.arn",
				TerraformMethod:   sdp.QueryMethod_SEARCH,
			},
		},
		PotentialLinks: []string{"ec2-launch-template", "elbv2-target-group", "ec2-instance", "iam-role", "autoscaling-launch-configuration", "ec2-placement-group"},
	}
}
