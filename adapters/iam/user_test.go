package iam

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func (t *TestIAMClient) ListGroupsForUser(ctx context.Context, params *iam.ListGroupsForUserInput, optFns ...func(*iam.Options)) (*iam.ListGroupsForUserOutput, error) {
	isTruncated := true
	marker := params.Marker

	if marker == nil {
		marker = adapters.PtrString("0")
	}

	// Get the current page
	markerInt, _ := strconv.Atoi(*marker)

	// Set the marker to the next page
	markerInt++

	if markerInt >= 3 {
		isTruncated = false
		marker = nil
	} else {
		marker = adapters.PtrString(fmt.Sprint(markerInt))
	}

	return &iam.ListGroupsForUserOutput{
		Groups: []types.Group{
			{
				Arn:        adapters.PtrString("arn:aws:iam::801795385023:Group/something"),
				CreateDate: adapters.PtrTime(time.Now()),
				GroupId:    adapters.PtrString("id"),
				GroupName:  adapters.PtrString(fmt.Sprintf("group-%v", marker)),
				Path:       adapters.PtrString("/"),
			},
		},
		IsTruncated: isTruncated,
		Marker:      marker,
	}, nil
}

func (t *TestIAMClient) GetUser(ctx context.Context, params *iam.GetUserInput, optFns ...func(*iam.Options)) (*iam.GetUserOutput, error) {
	return &iam.GetUserOutput{
		User: &types.User{
			Path:       adapters.PtrString("/"),
			UserName:   adapters.PtrString("power-users"),
			UserId:     adapters.PtrString("AGPA3VLV2U27T6SSLJMDS"),
			Arn:        adapters.PtrString("arn:aws:iam::801795385023:User/power-users"),
			CreateDate: adapters.PtrTime(time.Now()),
		},
	}, nil
}

func (t *TestIAMClient) ListUsers(ctx context.Context, params *iam.ListUsersInput, optFns ...func(*iam.Options)) (*iam.ListUsersOutput, error) {
	isTruncated := true
	marker := params.Marker

	if marker == nil {
		marker = adapters.PtrString("0")
	}

	// Get the current page
	markerInt, _ := strconv.Atoi(*marker)

	// Set the marker to the next page
	markerInt++

	if markerInt >= 3 {
		isTruncated = false
		marker = nil
	} else {
		marker = adapters.PtrString(fmt.Sprint(markerInt))
	}

	return &iam.ListUsersOutput{
		Users: []types.User{
			{
				Path:       adapters.PtrString("/"),
				UserName:   adapters.PtrString(fmt.Sprintf("user-%v", marker)),
				UserId:     adapters.PtrString("AGPA3VLV2U27T6SSLJMDS"),
				Arn:        adapters.PtrString("arn:aws:iam::801795385023:User/power-users"),
				CreateDate: adapters.PtrTime(time.Now()),
			},
		},
		IsTruncated: isTruncated,
		Marker:      marker,
	}, nil
}

func (t *TestIAMClient) ListUserTags(context.Context, *iam.ListUserTagsInput, ...func(*iam.Options)) (*iam.ListUserTagsOutput, error) {
	return &iam.ListUserTagsOutput{
		Tags: []types.Tag{
			{
				Key:   adapters.PtrString("foo"),
				Value: adapters.PtrString("bar"),
			},
		},
		IsTruncated: false,
		Marker:      nil,
	}, nil
}

func TestGetUserGroups(t *testing.T) {
	groups, err := getUserGroups(context.Background(), &TestIAMClient{}, adapters.PtrString("foo"))

	if err != nil {
		t.Error(err)
	}

	if len(groups) != 3 {
		t.Errorf("expected 3 groups, got %v", len(groups))
	}
}

func TestUserGetFunc(t *testing.T) {
	user, err := userGetFunc(context.Background(), &TestIAMClient{}, "foo", "bar")

	if err != nil {
		t.Error(err)
	}

	if user.User == nil {
		t.Error("user is nil")
	}

	if len(user.UserGroups) != 3 {
		t.Errorf("expected 3 groups, got %v", len(user.UserGroups))

	}
}

func TestUserListFunc(t *testing.T) {
	users, err := userListFunc(context.Background(), &TestIAMClient{}, "foo")

	if err != nil {
		t.Error(err)
	}

	if len(users) != 3 {
		t.Errorf("expected 3 users, got %v", len(users))
	}

	for _, user := range users {
		if len(user.UserGroups) != 3 {
			t.Errorf("expected 3 groups, got %v", len(user.UserGroups))
		}
	}
}

func TestUserListTagsFunc(t *testing.T) {
	tags, err := userListTagsFunc(context.Background(), &UserDetails{
		User: &types.User{
			UserName: adapters.PtrString("foo"),
		},
	}, &TestIAMClient{})

	if err != nil {
		t.Error(err)
	}

	if len(tags) != 1 {
		t.Errorf("expected 1 tag, got %v", len(tags))
	}
}

func TestUserItemMapper(t *testing.T) {
	details := UserDetails{
		User: &types.User{
			Path:       adapters.PtrString("/"),
			UserName:   adapters.PtrString("power-users"),
			UserId:     adapters.PtrString("AGPA3VLV2U27T6SSLJMDS"),
			Arn:        adapters.PtrString("arn:aws:iam::801795385023:User/power-users"),
			CreateDate: adapters.PtrTime(time.Now()),
		},
		UserGroups: []types.Group{
			{
				Arn:        adapters.PtrString("arn:aws:iam::801795385023:Group/something"),
				CreateDate: adapters.PtrTime(time.Now()),
				GroupId:    adapters.PtrString("id"),
				GroupName:  adapters.PtrString("name"),
				Path:       adapters.PtrString("/"),
			},
		},
	}

	item, err := userItemMapper("", "foo", &details)

	if err != nil {
		t.Error(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}

	tests := adapters.QueryTests{
		{
			ExpectedType:   "iam-group",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "name",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)
}

func TestNewUserSource(t *testing.T) {
	config, account, region := adapters.GetAutoConfig(t)
	client := iam.NewFromConfig(config, func(o *iam.Options) {
		o.RetryMode = aws.RetryModeAdaptive
		o.RetryMaxAttempts = 10
	})

	source := NewUserSource(client, account, region)

	test := adapters.E2ETest{
		Adapter: source,
		Timeout: 30 * time.Second,
	}

	test.Run(t)
}
