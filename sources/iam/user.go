package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"

	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

type UserDetails struct {
	User       *types.User
	UserGroups []types.Group
}

func UserGetFunc(ctx context.Context, client IAMClient, scope, query string) (*UserDetails, error) {
	out, err := client.GetUser(ctx, &iam.GetUserInput{
		UserName: &query,
	})

	if err != nil {
		return nil, err
	}

	details := UserDetails{
		User: out.User,
	}

	if out.User != nil {
		// Get the groups that the user is in too soe that we can create linked item requests
		groups, err := GetUserGroups(ctx, client, out.User.UserName)

		if err == nil {
			details.UserGroups = groups
		}
	}

	return &details, nil
}

// Gets all of the groups that a user is in
func GetUserGroups(ctx context.Context, client IAMClient, userName *string) ([]types.Group, error) {
	var out *iam.ListGroupsForUserOutput
	var err error
	groups := make([]types.Group, 0)

	paginator := iam.NewListGroupsForUserPaginator(client, &iam.ListGroupsForUserInput{
		UserName: userName,
	})

	for paginator.HasMorePages() {
		out, err = paginator.NextPage(ctx)

		if err != nil {
			return nil, err

		}

		groups = append(groups, out.Groups...)
	}

	return groups, nil
}

func UserListFunc(ctx context.Context, client IAMClient, scope string) ([]*UserDetails, error) {
	var out *iam.ListUsersOutput
	var err error
	users := make([]types.User, 0)

	paginator := iam.NewListUsersPaginator(client, &iam.ListUsersInput{})

	for paginator.HasMorePages() {
		out, err = paginator.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		users = append(users, out.Users...)
	}

	userDetails := make([]*UserDetails, len(users))

	for i, user := range users {
		details := UserDetails{
			User: &user,
		}

		groups, err := GetUserGroups(ctx, client, user.UserName)

		if err == nil {
			details.UserGroups = groups
		}

		userDetails[i] = &details
	}

	return userDetails, nil
}

func UserItemMapper(scope string, awsItem *UserDetails) (*sdp.Item, error) {
	attributes, err := sources.ToAttributesCase(awsItem.User)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "iam-user",
		UniqueAttribute: "userName",
		Attributes:      attributes,
		Scope:           scope,
	}

	for _, group := range awsItem.UserGroups {
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
			Type:   "iam-group",
			Method: sdp.QueryMethod_GET,
			Query:  *group.GroupName,
			Scope:  scope,
		})
	}

	return &item, nil
}

func NewUserSource(config aws.Config, accountID string, region string) *sources.GetListSource[*UserDetails, IAMClient, *iam.Options] {
	return &sources.GetListSource[*UserDetails, IAMClient, *iam.Options]{
		ItemType:   "iam-user",
		Client:     iam.NewFromConfig(config),
		AccountID:  accountID,
		Region:     region,
		GetFunc:    UserGetFunc,
		ListFunc:   UserListFunc,
		ItemMapper: UserItemMapper,
	}
}
