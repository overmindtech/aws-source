package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func SecurityGroupInputMapperGet(scope string, query string) (*ec2.DescribeSecurityGroupsInput, error) {
	return &ec2.DescribeSecurityGroupsInput{
		GroupIds: []string{
			query,
		},
	}, nil
}

func SecurityGroupInputMapperList(scope string) (*ec2.DescribeSecurityGroupsInput, error) {
	return &ec2.DescribeSecurityGroupsInput{}, nil
}

func SecurityGroupOutputMapper(scope string, output *ec2.DescribeSecurityGroupsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, securityGroup := range output.SecurityGroups {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(securityGroup)

		if err != nil {
			return nil, &sdp.ItemRequestError{
				ErrorType:   sdp.ItemRequestError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		item := sdp.Item{
			Type:            "ec2-security-group",
			UniqueAttribute: "groupId",
			Scope:           scope,
			Attributes:      attrs,
		}

		// VPC
		if securityGroup.VpcId != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "ec2-vpc",
				Method: sdp.RequestMethod_GET,
				Query:  *securityGroup.VpcId,
				Scope:  scope,
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

func NewSecurityGroupSource(config aws.Config, accountID string, limit *LimitBucket) *sources.DescribeOnlySource[*ec2.DescribeSecurityGroupsInput, *ec2.DescribeSecurityGroupsOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeSecurityGroupsInput, *ec2.DescribeSecurityGroupsOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		AccountID: accountID,
		ItemType:  "ec2-security-group",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error) {
			<-limit.C // Wait for late limiting
			return client.DescribeSecurityGroups(ctx, input)
		},
		InputMapperGet:  SecurityGroupInputMapperGet,
		InputMapperList: SecurityGroupInputMapperList,
		PaginatorBuilder: func(client *ec2.Client, params *ec2.DescribeSecurityGroupsInput) sources.Paginator[*ec2.DescribeSecurityGroupsOutput, *ec2.Options] {
			return ec2.NewDescribeSecurityGroupsPaginator(client, params)
		},
		OutputMapper: SecurityGroupOutputMapper,
	}
}
