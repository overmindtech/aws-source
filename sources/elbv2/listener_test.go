package elbv2

import (
	"testing"

	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestListenerOutputMapper(t *testing.T) {
	output := elbv2.DescribeListenersOutput{
		Listeners: []types.Listener{
			{
				ListenerArn:     sources.PtrString("arn:aws:elasticloadbalancing:eu-west-2:944651592624:listener/app/ingress/1bf10920c5bd199d/9d28f512be129134"),
				LoadBalancerArn: sources.PtrString("arn:aws:elasticloadbalancing:eu-west-2:944651592624:loadbalancer/app/ingress/1bf10920c5bd199d"), // link
				Port:            sources.PtrInt32(443),
				Protocol:        types.ProtocolEnumHttps,
				Certificates: []types.Certificate{
					{
						CertificateArn: sources.PtrString("arn:aws:acm:eu-west-2:944651592624:certificate/acd84d34-fb78-4411-bd8a-43684a3477c5"), // link
						IsDefault:      sources.PtrBool(true),
					},
				},
				SslPolicy: sources.PtrString("ELBSecurityPolicy-2016-08"),
				AlpnPolicy: []string{
					"policy1",
				},
				DefaultActions: []types.Action{
					{
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
					},
				},
			},
		},
	}

	items, err := ListenerOutputMapper("foo", &output)

	if err != nil {
		t.Error(err)
	}

	for _, item := range items {
		if err := item.Validate(); err != nil {
			t.Error(err)
		}
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %v", len(items))
	}

	item := items[0]

	// It doesn't really make sense to test anything other than the linked items
	// since the attributes are converted automatically
	tests := sources.ItemRequestTests{
		{
			ExpectedType:   "elbv2-load-balancer",
			ExpectedMethod: sdp.RequestMethod_SEARCH,
			ExpectedQuery:  "arn:aws:elasticloadbalancing:eu-west-2:944651592624:loadbalancer/app/ingress/1bf10920c5bd199d",
			ExpectedScope:  "944651592624.eu-west-2",
		},
		{
			ExpectedType:   "acm-certificate",
			ExpectedMethod: sdp.RequestMethod_SEARCH,
			ExpectedQuery:  "arn:aws:acm:eu-west-2:944651592624:certificate/acd84d34-fb78-4411-bd8a-43684a3477c5",
			ExpectedScope:  "944651592624.eu-west-2",
		},
		{
			ExpectedType:   "cognito-idp-user-pool",
			ExpectedMethod: sdp.RequestMethod_SEARCH,
			ExpectedQuery:  "arn:partition:service:region:account-id:resource-type:resource-id",
			ExpectedScope:  "account-id.region",
		},
		{
			ExpectedType:   "http",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "https://auth.somewhere.com/app1",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "http",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "https://auth.somewhere.com/app1/tokens",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "http",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "https://auth.somewhere.com/app1/users",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "elbv2-target-group",
			ExpectedMethod: sdp.RequestMethod_SEARCH,
			ExpectedQuery:  "arn:partition:service:region:account-id:resource-type:resource-id",
			ExpectedScope:  "account-id.region",
		},
		{
			ExpectedType:   "http",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "https://somewhere.else.com:8080/login?foo=bar",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "elbv2-target-group",
			ExpectedMethod: sdp.RequestMethod_SEARCH,
			ExpectedQuery:  "arn:partition:service:region:account-id:resource-type:resource-id",
			ExpectedScope:  "account-id.region",
		},
	}

	tests.Execute(t, item)
}
