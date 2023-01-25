package eks

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func NodegroupGetFunc(ctx context.Context, client EKSClient, scope string, input *eks.DescribeNodegroupInput) (*sdp.Item, error) {
	out, err := client.DescribeNodegroup(ctx, input)

	if err != nil {
		return nil, err
	}

	if out.Nodegroup == nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOTFOUND,
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
	}

	if ng.RemoteAccess != nil {
		if ng.RemoteAccess.Ec2SshKey != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "ec2-key-pair",
				Method: sdp.RequestMethod_GET,
				Query:  *ng.RemoteAccess.Ec2SshKey,
				Scope:  scope,
			})
		}

		for _, sg := range ng.RemoteAccess.SourceSecurityGroups {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "ec2-security-group",
				Method: sdp.RequestMethod_GET,
				Query:  sg,
				Scope:  scope,
			})
		}
	}

	for _, subnet := range ng.Subnets {
		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Type:   "ec2-subnet",
			Method: sdp.RequestMethod_GET,
			Query:  subnet,
			Scope:  scope,
		})
	}

	if ng.Resources != nil {
		for _, g := range ng.Resources.AutoScalingGroups {
			if g.Name != nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "autoscaling-auto-scaling-group",
					Method: sdp.RequestMethod_GET,
					Query:  *g.Name,
					Scope:  scope,
				})
			}
		}

		if ng.Resources.RemoteAccessSecurityGroup != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "ec2-security-group",
				Method: sdp.RequestMethod_GET,
				Query:  *ng.Resources.RemoteAccessSecurityGroup,
				Scope:  scope,
			})
		}
	}

	if ng.LaunchTemplate != nil {
		if ng.LaunchTemplate.Id != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "ec2-launch-template",
				Method: sdp.RequestMethod_GET,
				Query:  *ng.LaunchTemplate.Id,
				Scope:  scope,
			})
		}
	}

	return &item, nil
}

func NewNodegroupSource(config aws.Config, accountID string, region string) *sources.AlwaysGetSource[*eks.ListNodegroupsInput, *eks.ListNodegroupsOutput, *eks.DescribeNodegroupInput, *eks.DescribeNodegroupOutput, EKSClient, *eks.Options] {
	return &sources.AlwaysGetSource[*eks.ListNodegroupsInput, *eks.ListNodegroupsOutput, *eks.DescribeNodegroupInput, *eks.DescribeNodegroupOutput, EKSClient, *eks.Options]{
		ItemType:    "eks-nodegroup",
		Client:      eks.NewFromConfig(config),
		AccountID:   accountID,
		Region:      region,
		DisableList: true,
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

			if len(fields) == 2 {
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
			inputs := make([]*eks.DescribeNodegroupInput, len(output.Nodegroups))

			for i, group := range output.Nodegroups {
				inputs[i] = &eks.DescribeNodegroupInput{
					ClusterName:   input.ClusterName,
					NodegroupName: &group,
				}
			}

			return inputs, nil
		},
		GetFunc: NodegroupGetFunc,
	}
}
