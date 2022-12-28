package eks

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func ClusterGetFunc(ctx context.Context, client EKSClient, scope string, input *eks.DescribeClusterInput) (*sdp.Item, error) {
	output, err := client.DescribeCluster(ctx, input)

	if err != nil {
		return nil, err
	}

	if output.Cluster == nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOTFOUND,
			ErrorString: "cluster response was nil",
		}
	}

	cluster := output.Cluster

	attributes, err := sources.ToAttributesCase(cluster, "clientRequestToken")

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "eks-cluster",
		UniqueAttribute: "name",
		Attributes:      attributes,
		Scope:           scope,
		LinkedItemRequests: []*sdp.ItemRequest{
			{
				Type:   "eks-addon",
				Method: sdp.RequestMethod_SEARCH,
				Query:  *cluster.Name,
				Scope:  scope,
			},
			{
				Type:   "eks-fargate-profile",
				Method: sdp.RequestMethod_SEARCH,
				Query:  *cluster.Name,
				Scope:  scope,
			},
			{
				Type:   "eks-nodegroup",
				Method: sdp.RequestMethod_SEARCH,
				Query:  *cluster.Name,
				Scope:  scope,
			},
		},
	}

	var a *sources.ARN

	if cluster.ConnectorConfig != nil {
		if cluster.ConnectorConfig.RoleArn != nil {
			if a, err = sources.ParseARN(*cluster.ConnectorConfig.RoleArn); err == nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "iam-role",
					Method: sdp.RequestMethod_SEARCH,
					Query:  *cluster.ConnectorConfig.RoleArn,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				})
			}
		}
	}

	for _, conf := range cluster.EncryptionConfig {
		if conf.Provider != nil {
			if conf.Provider.KeyArn != nil {
				if a, err = sources.ParseARN(*conf.Provider.KeyArn); err == nil {
					item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
						Type:   "kms-key",
						Method: sdp.RequestMethod_SEARCH,
						Query:  *conf.Provider.KeyArn,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					})
				}
			}
		}
	}

	if cluster.Endpoint != nil {
		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Type:   "http",
			Method: sdp.RequestMethod_GET,
			Query:  *cluster.Endpoint,
			Scope:  "global",
		})
	}

	if cluster.ResourcesVpcConfig != nil {
		if cluster.ResourcesVpcConfig.ClusterSecurityGroupId != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "ec2-security-group",
				Method: sdp.RequestMethod_GET,
				Query:  *cluster.ResourcesVpcConfig.ClusterSecurityGroupId,
				Scope:  scope,
			})
		}

		for _, id := range cluster.ResourcesVpcConfig.SecurityGroupIds {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "ec2-security-group",
				Method: sdp.RequestMethod_GET,
				Query:  id,
				Scope:  scope,
			})
		}

		for _, id := range cluster.ResourcesVpcConfig.SubnetIds {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "ec2-subnet",
				Method: sdp.RequestMethod_GET,
				Query:  id,
				Scope:  scope,
			})
		}

		if cluster.ResourcesVpcConfig.VpcId != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "ec2-vpc",
				Method: sdp.RequestMethod_GET,
				Query:  *cluster.ResourcesVpcConfig.VpcId,
				Scope:  scope,
			})
		}
	}

	if cluster.RoleArn != nil {
		if a, err = sources.ParseARN(*cluster.RoleArn); err == nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "iam-role",
				Method: sdp.RequestMethod_SEARCH,
				Query:  *cluster.RoleArn,
				Scope:  sources.FormatScope(a.AccountID, a.Region),
			})
		}
	}

	return &item, nil

}

func NewClusterSource(config aws.Config, accountID string, region string) *sources.ListGetSource[*eks.ListClustersInput, *eks.ListClustersOutput, *eks.DescribeClusterInput, *eks.DescribeClusterOutput, EKSClient, *eks.Options] {
	return &sources.ListGetSource[*eks.ListClustersInput, *eks.ListClustersOutput, *eks.DescribeClusterInput, *eks.DescribeClusterOutput, EKSClient, *eks.Options]{
		ItemType:  "eks-cluster",
		Client:    eks.NewFromConfig(config),
		AccountID: accountID,
		Region:    region,
		ListInput: &eks.ListClustersInput{},
		GetInputMapper: func(scope, query string) *eks.DescribeClusterInput {
			return &eks.DescribeClusterInput{
				Name: &query,
			}
		},
		ListFuncPaginatorBuilder: func(client EKSClient, input *eks.ListClustersInput) sources.Paginator[*eks.ListClustersOutput, *eks.Options] {
			return eks.NewListClustersPaginator(client, input)
		},
		ListFuncOutputMapper: func(output *eks.ListClustersOutput) ([]*eks.DescribeClusterInput, error) {
			inputs := make([]*eks.DescribeClusterInput, len(output.Clusters))

			for i, clusterName := range output.Clusters {
				inputs[i] = &eks.DescribeClusterInput{
					Name: &clusterName,
				}
			}

			return inputs, nil
		},
		GetFunc: ClusterGetFunc,
	}
}
