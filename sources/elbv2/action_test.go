package elbv2

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestActionToRequests(t *testing.T) {
	action := types.Action{
		Type:  types.ActionTypeEnumFixedResponse,
		Order: sources.PtrInt32(1),
		FixedResponseConfig: &types.FixedResponseActionConfig{
			StatusCode:  sources.PtrString("404"),
			ContentType: sources.PtrString("text/plain"),
			MessageBody: sources.PtrString("not found"),
		},
		AuthenticateCognitoConfig: &types.AuthenticateCognitoActionConfig{
			UserPoolArn:      sources.PtrString("arn:partition:service:region:account-id:resource-type:resource-id"), // link
			UserPoolClientId: sources.PtrString("clientID"),
			UserPoolDomain:   sources.PtrString("domain.com"),
			AuthenticationRequestExtraParams: map[string]string{
				"foo": "bar",
			},
			OnUnauthenticatedRequest: types.AuthenticateCognitoActionConditionalBehaviorEnumAuthenticate,
			Scope:                    sources.PtrString("foo"),
			SessionCookieName:        sources.PtrString("cookie"),
			SessionTimeout:           sources.PtrInt64(10),
		},
		AuthenticateOidcConfig: &types.AuthenticateOidcActionConfig{
			AuthorizationEndpoint:            sources.PtrString("https://auth.somewhere.com/app1"), // link
			ClientId:                         sources.PtrString("CLIENT-ID"),
			Issuer:                           sources.PtrString("Someone"),
			TokenEndpoint:                    sources.PtrString("https://auth.somewhere.com/app1/tokens"), // link
			UserInfoEndpoint:                 sources.PtrString("https://auth.somewhere.com/app1/users"),  // link
			AuthenticationRequestExtraParams: map[string]string{},
			ClientSecret:                     sources.PtrString("secret"), // Redact
			OnUnauthenticatedRequest:         types.AuthenticateOidcActionConditionalBehaviorEnumAllow,
			Scope:                            sources.PtrString("foo"),
			SessionCookieName:                sources.PtrString("cookie"),
			SessionTimeout:                   sources.PtrInt64(10),
			UseExistingClientSecret:          sources.PtrBool(true),
		},
		ForwardConfig: &types.ForwardActionConfig{
			TargetGroupStickinessConfig: &types.TargetGroupStickinessConfig{
				DurationSeconds: sources.PtrInt32(10),
				Enabled:         sources.PtrBool(true),
			},
			TargetGroups: []types.TargetGroupTuple{
				{
					TargetGroupArn: sources.PtrString("arn:partition:service:region:account-id:resource-type:resource-id"), // link
					Weight:         sources.PtrInt32(1),
				},
			},
		},
		RedirectConfig: &types.RedirectActionConfig{
			StatusCode: types.RedirectActionStatusCodeEnumHttp302,
			Host:       sources.PtrString("somewhere.else.com"), // combine and link
			Path:       sources.PtrString("/login"),             // combine and link
			Port:       sources.PtrString("8080"),               // combine and link
			Protocol:   sources.PtrString("https"),              // combine and link
			Query:      sources.PtrString("foo=bar"),            // combine and link
		},
		TargetGroupArn: sources.PtrString("arn:partition:service:region:account-id:resource-type:resource-id"), // link
	}

	item := sdp.Item{
		Type:              "test",
		UniqueAttribute:   "foo",
		Attributes:        &sdp.ItemAttributes{},
		Scope:             "foo",
		LinkedItemQueries: ActionToRequests(action),
	}

	tests := sources.QueryTests{
		{
			ExpectedType:   "cognito-idp-user-pool",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:partition:service:region:account-id:resource-type:resource-id",
			ExpectedScope:  "account-id.region",
		},
		{
			ExpectedType:   "http",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "https://auth.somewhere.com/app1",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "http",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "https://auth.somewhere.com/app1/tokens",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "http",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "https://auth.somewhere.com/app1/users",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "elbv2-target-group",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:partition:service:region:account-id:resource-type:resource-id",
			ExpectedScope:  "account-id.region",
		},
		{
			ExpectedType:   "http",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "https://somewhere.else.com:8080/login?foo=bar",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "elbv2-target-group",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:partition:service:region:account-id:resource-type:resource-id",
			ExpectedScope:  "account-id.region",
		},
	}

	tests.Execute(t, &item)
}
