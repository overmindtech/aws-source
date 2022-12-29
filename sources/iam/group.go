package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"

	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func GroupGetFunc(ctx context.Context, client *iam.Client, scope, query string) (*types.Group, error) {
	out, err := client.GetGroup(ctx, &iam.GetGroupInput{
		GroupName: &query,
	})

	if err != nil {
		return nil, err
	}

	return out.Group, nil
}

func GroupListFunc(ctx context.Context, client *iam.Client, scope string) ([]*types.Group, error) {
	out, err := client.ListGroups(ctx, &iam.ListGroupsInput{})

	if err != nil {
		return nil, err
	}

	zones := make([]*types.Group, len(out.Groups))

	for i, zone := range out.Groups {
		zones[i] = &zone
	}

	return zones, nil
}

func GroupItemMapper(scope string, awsItem *types.Group) (*sdp.Item, error) {
	attributes, err := sources.ToAttributesCase(awsItem)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "iam-group",
		UniqueAttribute: "groupName",
		Attributes:      attributes,
		Scope:           scope,
	}

	return &item, nil
}

func NewGroupSource(config aws.Config, accountID string, region string) *sources.GetListSource[*types.Group, *iam.Client, *iam.Options] {
	return &sources.GetListSource[*types.Group, *iam.Client, *iam.Options]{
		ItemType:   "iam-group",
		Client:     iam.NewFromConfig(config),
		AccountID:  accountID,
		Region:     region,
		GetFunc:    GroupGetFunc,
		ListFunc:   GroupListFunc,
		ItemMapper: GroupItemMapper,
	}
}
