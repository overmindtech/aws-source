package elbv2

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func TestActionToRequests(t *testing.T) {
	action := types.Action{
		Type:  types.ActionTypeEnumFixedResponse,
		Order: adapters.PtrInt32(1),
		FixedResponseConfig: &types.FixedResponseActionConfig{
			StatusCode:  adapters.PtrString("404"),
			ContentType: adapters.PtrString("text/plain"),
			MessageBody: adapters.PtrString("not found"),
		},
		AuthenticateCognitoConfig: &types.AuthenticateCognitoActionConfig{
			UserPoolArn:      adapters.PtrString("arn:partition:service:region:account-id:resource-type:resource-id"), // link
			UserPoolClientId: adapters.PtrString("clientID"),
			UserPoolDomain:   adapters.PtrString("domain.com"),
			AuthenticationRequestExtraParams: map[string]string{
				"foo": "bar",
			},
			OnUnauthenticatedRequest: types.AuthenticateCognitoActionConditionalBehaviorEnumAuthenticate,
			Scope:                    adapters.PtrString("foo"),
			SessionCookieName:        adapters.PtrString("cookie"),
			SessionTimeout:           adapters.PtrInt64(10),
		},
		AuthenticateOidcConfig: &types.AuthenticateOidcActionConfig{
			AuthorizationEndpoint:            adapters.PtrString("https://auth.somewhere.com/app1"), // link
			ClientId:                         adapters.PtrString("CLIENT-ID"),
			Issuer:                           adapters.PtrString("Someone"),
			TokenEndpoint:                    adapters.PtrString("https://auth.somewhere.com/app1/tokens"), // link
			UserInfoEndpoint:                 adapters.PtrString("https://auth.somewhere.com/app1/users"),  // link
			AuthenticationRequestExtraParams: map[string]string{},
			ClientSecret:                     adapters.PtrString("secret"), // Redact
			OnUnauthenticatedRequest:         types.AuthenticateOidcActionConditionalBehaviorEnumAllow,
			Scope:                            adapters.PtrString("foo"),
			SessionCookieName:                adapters.PtrString("cookie"),
			SessionTimeout:                   adapters.PtrInt64(10),
			UseExistingClientSecret:          adapters.PtrBool(true),
		},
		ForwardConfig: &types.ForwardActionConfig{
			TargetGroupStickinessConfig: &types.TargetGroupStickinessConfig{
				DurationSeconds: adapters.PtrInt32(10),
				Enabled:         adapters.PtrBool(true),
			},
			TargetGroups: []types.TargetGroupTuple{
				{
					TargetGroupArn: adapters.PtrString("arn:partition:service:region:account-id:resource-type:resource-id1"), // link
					Weight:         adapters.PtrInt32(1),
				},
			},
		},
		RedirectConfig: &types.RedirectActionConfig{
			StatusCode: types.RedirectActionStatusCodeEnumHttp302,
			Host:       adapters.PtrString("somewhere.else.com"), // combine and link
			Path:       adapters.PtrString("/login"),             // combine and link
			Port:       adapters.PtrString("8080"),               // combine and link
			Protocol:   adapters.PtrString("https"),              // combine and link
			Query:      adapters.PtrString("foo=bar"),            // combine and link
		},
		TargetGroupArn: adapters.PtrString("arn:partition:service:region:account-id:resource-type:resource-id2"), // link
	}

	item := sdp.Item{
		Type:              "test",
		UniqueAttribute:   "foo",
		Attributes:        &sdp.ItemAttributes{},
		Scope:             "foo",
		LinkedItemQueries: ActionToRequests(action),
	}

	tests := adapters.QueryTests{
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
			ExpectedQuery:  "arn:partition:service:region:account-id:resource-type:resource-id1",
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
			ExpectedQuery:  "arn:partition:service:region:account-id:resource-type:resource-id2",
			ExpectedScope:  "account-id.region",
		},
	}

	tests.Execute(t, &item)
}
