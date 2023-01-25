package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
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

		item.LinkedItemRequests = append(item.LinkedItemRequests, extractLinkedSecurityGroups(securityGroup.IpPermissions, scope)...)
		item.LinkedItemRequests = append(item.LinkedItemRequests, extractLinkedSecurityGroups(securityGroup.IpPermissionsEgress, scope)...)

		items = append(items, &item)
	}

	return items, nil
}

func NewSecurityGroupSource(config aws.Config, accountID string, limit *LimitBucket) *sources.DescribeOnlySource[*ec2.DescribeSecurityGroupsInput, *ec2.DescribeSecurityGroupsOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeSecurityGroupsInput, *ec2.DescribeSecurityGroupsOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		Client:    ec2.NewFromConfig(config),
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

// extractLinkedSecurityGroups Extracts related security groups from IP
// permissions
func extractLinkedSecurityGroups(permissions []types.IpPermission, scope string) []*sdp.ItemRequest {
	currentAccount, region, err := sources.ParseScope(scope)
	requests := make([]*sdp.ItemRequest, 0)
	var relatedAccount string

	if err != nil {
		return requests
	}

	for _, permission := range permissions {
		for _, idGroup := range permission.UserIdGroupPairs {
			if idGroup.UserId != nil {
				relatedAccount = *idGroup.UserId
			} else {
				relatedAccount = currentAccount
			}

			if idGroup.GroupId != nil {
				requests = append(requests, &sdp.ItemRequest{
					Type:   "ec2-security-group",
					Method: sdp.RequestMethod_GET,
					Query:  *idGroup.GroupId,
					Scope:  sources.FormatScope(relatedAccount, region),
				})
			}
		}
	}

	return requests
}
