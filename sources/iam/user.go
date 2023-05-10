package iam

import (
	"context"
	"time"

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

func userGetFunc(ctx context.Context, client IAMClient, scope, query string, limit *sources.LimitBucket) (*UserDetails, error) {
	<-limit.C
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
		enrichUser(ctx, client, &details, limit)
	}

	return &details, nil
}

// enrichUser Enriches the user with group and tag info
func enrichUser(ctx context.Context, client IAMClient, userDetails *UserDetails, limit *sources.LimitBucket) error {
	var err error

	userDetails.UserGroups, err = getUserGroups(ctx, client, userDetails.User.UserName, limit)

	if err != nil {
		return err
	}

	userDetails.User.Tags, err = getUserTags(ctx, client, userDetails.User.UserName, limit)

	if err != nil {
		return err
	}

	return nil
}

// Gets all of the groups that a user is in
func getUserGroups(ctx context.Context, client IAMClient, userName *string, limit *sources.LimitBucket) ([]types.Group, error) {
	var out *iam.ListGroupsForUserOutput
	var err error
	groups := make([]types.Group, 0)

	paginator := iam.NewListGroupsForUserPaginator(client, &iam.ListGroupsForUserInput{
		UserName: userName,
	})

	for paginator.HasMorePages() {
		<-limit.C
		out, err = paginator.NextPage(ctx)

		if err != nil {
			return nil, err

		}

		groups = append(groups, out.Groups...)
	}

	return groups, nil
}

// GetUserTags Gets the tags for a user since the API doesn't actually return
// this, even though it says it does see:
// https://github.com/boto/boto3/issues/1855
func getUserTags(ctx context.Context, client IAMClient, userName *string, limit *sources.LimitBucket) ([]types.Tag, error) {
	paginator := iam.NewListUserTagsPaginator(client, &iam.ListUserTagsInput{
		UserName: userName,
	})

	var out *iam.ListUserTagsOutput
	var err error

	tags := make([]types.Tag, 0)

	for paginator.HasMorePages() {
		<-limit.C
		out, err = paginator.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		tags = append(tags, out.Tags...)
	}

	return tags, err
}

func userListFunc(ctx context.Context, client IAMClient, scope string, limit *sources.LimitBucket) ([]*UserDetails, error) {
	var out *iam.ListUsersOutput
	var err error
	users := make([]types.User, 0)

	paginator := iam.NewListUsersPaginator(client, &iam.ListUsersInput{})

	for paginator.HasMorePages() {
		<-limit.C
		out, err = paginator.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		users = append(users, out.Users...)
	}

	userDetails := make([]*UserDetails, len(users))

	for i := range users {
		details := UserDetails{
			User: &users[i],
		}

		enrichUser(ctx, client, &details, limit)

		userDetails[i] = &details
	}

	return userDetails, nil
}

func userItemMapper(scope string, awsItem *UserDetails) (*sdp.Item, error) {
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
		// +overmind:link iam-group
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{Query: &sdp.Query{
			Type:   "iam-group",
			Method: sdp.QueryMethod_GET,
			Query:  *group.GroupName,
			Scope:  scope,
		}})
	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type iam-user
// +overmind:descriptiveType IAM User
// +overmind:get Get a user by name
// +overmind:list List all users
// +overmind:search Search for users by ARN
// +overmind:group AWS

func NewUserSource(config aws.Config, accountID string, region string, limit *sources.LimitBucket) *sources.GetListSource[*UserDetails, IAMClient, *iam.Options] {
	return &sources.GetListSource[*UserDetails, IAMClient, *iam.Options]{
		ItemType:      "iam-user",
		Client:        iam.NewFromConfig(config),
		AccountID:     accountID,
		CacheDuration: 1 * time.Hour, // IAM has very low rate limits, we need to cache for a long time
		Region:        region,
		GetFunc: func(ctx context.Context, client IAMClient, scope, query string) (*UserDetails, error) {
			return userGetFunc(ctx, client, scope, query, limit)
		},
		ListFunc: func(ctx context.Context, client IAMClient, scope string) ([]*UserDetails, error) {
			return userListFunc(ctx, client, scope, limit)
		},
		ItemMapper: userItemMapper,
	}
}
