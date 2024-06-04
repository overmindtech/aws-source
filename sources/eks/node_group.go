package eks

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func nodegroupGetFunc(ctx context.Context, client EKSClient, scope string, input *eks.DescribeNodegroupInput) (*sdp.Item, error) {
	out, err := client.DescribeNodegroup(ctx, input)

	if err != nil {
		return nil, err
	}

	if out.Nodegroup == nil {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOTFOUND,
			ErrorString: "Nodegroup was nil",
		}
	}

	attributes, err := sources.ToAttributesCase(out.Nodegroup)

	if err != nil {
		return nil, err
	}

	ng := out.Nodegroup

	// The uniqueAttributeValue for this is a custom field:
	// {clusterName}/{NodegroupName}
	attributes.Set("uniqueName", (*out.Nodegroup.ClusterName + "/" + *out.Nodegroup.NodegroupName))

	item := sdp.Item{
		Type:            "eks-nodegroup",
		UniqueAttribute: "uniqueName",
		Attributes:      attributes,
		Scope:           scope,
		Tags:            out.Nodegroup.Tags,
	}

	if ng.Health != nil {
		if len(ng.Health.Issues) > 0 {
			item.Health = sdp.Health_HEALTH_ERROR.Enum()
		} else {
			item.Health = sdp.Health_HEALTH_OK.Enum()
		}

		// NOTE: It would be good if we could link to the resource if there is a
		// health issue, but I can't find any examples of the format that the
		// `ResourceIds` array is in. If someone can find one, please add it here.
	}

	if ng.RemoteAccess != nil {
		if ng.RemoteAccess.Ec2SshKey != nil {
			// +overmind:link ec2-key-pair
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ec2-key-pair",
					Method: sdp.QueryMethod_GET,
					Query:  *ng.RemoteAccess.Ec2SshKey,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// The key pair can affect the node group
					In: true,
					// The node group can't affect the key pair
					Out: false,
				},
			})
		}

		for _, sg := range ng.RemoteAccess.SourceSecurityGroups {
			// +overmind:link ec2-security-group
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ec2-security-group",
					Method: sdp.QueryMethod_GET,
					Query:  sg,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// The security group can affect the node group
					In: true,
					// The node group can't affect the security group
					Out: false,
				},
			})
		}
	}

	for _, subnet := range ng.Subnets {
		// +overmind:link ec2-subnet
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "ec2-subnet",
				Method: sdp.QueryMethod_GET,
				Query:  subnet,
				Scope:  scope,
			},
			BlastPropagation: &sdp.BlastPropagation{
				// The subnet can affect the node group
				In: true,
				// The node group can't affect the subnet
				Out: false,
			},
		})
	}

	if ng.Resources != nil {
		for _, g := range ng.Resources.AutoScalingGroups {
			if g.Name != nil {
				// +overmind:link autoscaling-auto-scaling-group
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "autoscaling-auto-scaling-group",
						Method: sdp.QueryMethod_GET,
						Query:  *g.Name,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// These are tightly coupled
						In:  true,
						Out: true,
					},
				})
			}
		}

		if ng.Resources.RemoteAccessSecurityGroup != nil {
			// +overmind:link ec2-security-group
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ec2-security-group",
					Method: sdp.QueryMethod_GET,
					Query:  *ng.Resources.RemoteAccessSecurityGroup,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// The security group can affect the node group
					In: true,
					// The node group can't affect the security group
					Out: false,
				},
			})
		}
	}

	if ng.LaunchTemplate != nil {
		if ng.LaunchTemplate.Id != nil {
			// +overmind:link ec2-launch-template
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "ec2-launch-template",
					Method: sdp.QueryMethod_GET,
					Query:  *ng.LaunchTemplate.Id,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// The launch template can affect the node group
					In: true,
					// The node group can't affect the launch template
					Out: false,
				},
			})
		}
	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type eks-nodegroup
// +overmind:descriptiveType EKS Nodegroup
// +overmind:get Get a node group by unique name ({clusterName}/{NodegroupName})
// +overmind:list List all node groups
// +overmind:search Search for node groups by cluster name
// +overmind:group AWS
// +overmind:terraform:queryMap aws_eks_node_group.arn
// +overmind:terraform:method SEARCH

func NewNodegroupSource(client EKSClient, accountID string, region string) *sources.AlwaysGetSource[*eks.ListNodegroupsInput, *eks.ListNodegroupsOutput, *eks.DescribeNodegroupInput, *eks.DescribeNodegroupOutput, EKSClient, *eks.Options] {
	return &sources.AlwaysGetSource[*eks.ListNodegroupsInput, *eks.ListNodegroupsOutput, *eks.DescribeNodegroupInput, *eks.DescribeNodegroupOutput, EKSClient, *eks.Options]{
		ItemType:         "eks-nodegroup",
		Client:           client,
		AccountID:        accountID,
		Region:           region,
		DisableList:      true,
		AlwaysSearchARNs: true,
		SearchInputMapper: func(scope, query string) (*eks.ListNodegroupsInput, error) {
			return &eks.ListNodegroupsInput{
				ClusterName: &query,
			}, nil
		},
		GetInputMapper: func(scope, query string) *eks.DescribeNodegroupInput {
			// The uniqueAttributeValue for this is a custom field:
			// {clusterName}/{nodegroupName}
			fields := strings.Split(query, "/")

			var clusterName string
			var nodegroupName string

			if len(fields) >= 2 {
				clusterName = fields[0]
				nodegroupName = fields[1]
			}

			return &eks.DescribeNodegroupInput{
				NodegroupName: &nodegroupName,
				ClusterName:   &clusterName,
			}
		},
		ListFuncPaginatorBuilder: func(client EKSClient, input *eks.ListNodegroupsInput) sources.Paginator[*eks.ListNodegroupsOutput, *eks.Options] {
			return eks.NewListNodegroupsPaginator(client, input)
		},
		ListFuncOutputMapper: func(output *eks.ListNodegroupsOutput, input *eks.ListNodegroupsInput) ([]*eks.DescribeNodegroupInput, error) {
			inputs := make([]*eks.DescribeNodegroupInput, 0, len(output.Nodegroups))

			for i := range output.Nodegroups {
				inputs = append(inputs, &eks.DescribeNodegroupInput{
					ClusterName:   input.ClusterName,
					NodegroupName: &output.Nodegroups[i],
				})
			}

			return inputs, nil
		},
		GetFunc: nodegroupGetFunc,
	}
}
