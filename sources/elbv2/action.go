package elbv2

import (
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func ActionToRequests(action types.Action) []*sdp.LinkedItemQuery {
	requests := make([]*sdp.LinkedItemQuery, 0)

	if action.AuthenticateCognitoConfig != nil {
		if action.AuthenticateCognitoConfig.UserPoolArn != nil {
			if a, err := sources.ParseARN(*action.AuthenticateCognitoConfig.UserPoolArn); err == nil {
				requests = append(requests, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "cognito-idp-user-pool",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *action.AuthenticateCognitoConfig.UserPoolArn,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changing the user pool could affect the LB
						In: true,
						// The LB won't affect the user pool
						Out: false,
					},
				})
			}
		}
	}

	if action.AuthenticateOidcConfig != nil {
		if action.AuthenticateOidcConfig.AuthorizationEndpoint != nil {
			requests = append(requests, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "http",
					Method: sdp.QueryMethod_GET,
					Query:  *action.AuthenticateOidcConfig.AuthorizationEndpoint,
					Scope:  "global",
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changing the authorization endpoint could affect the LB
					In: true,
					// The LB won't affect the authorization endpoint
					Out: false,
				},
			})
		}

		if action.AuthenticateOidcConfig.TokenEndpoint != nil {
			requests = append(requests, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "http",
					Method: sdp.QueryMethod_GET,
					Query:  *action.AuthenticateOidcConfig.TokenEndpoint,
					Scope:  "global",
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changing the authorization endpoint could affect the LB
					In: true,
					// The LB won't affect the authorization endpoint
					Out: false,
				},
			})
		}

		if action.AuthenticateOidcConfig.UserInfoEndpoint != nil {
			requests = append(requests, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "http",
					Method: sdp.QueryMethod_GET,
					Query:  *action.AuthenticateOidcConfig.UserInfoEndpoint,
					Scope:  "global",
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Changing the authorization endpoint could affect the LB
					In: true,
					// The LB won't affect the authorization endpoint
					Out: false,
				},
			})
		}

		if action.ForwardConfig != nil {
			for _, tg := range action.ForwardConfig.TargetGroups {
				if tg.TargetGroupArn != nil {
					if a, err := sources.ParseARN(*tg.TargetGroupArn); err == nil {
						requests = append(requests, &sdp.LinkedItemQuery{
							Query: &sdp.Query{
								Type:   "elbv2-target-group",
								Method: sdp.QueryMethod_SEARCH,
								Query:  *tg.TargetGroupArn,
								Scope:  sources.FormatScope(a.AccountID, a.Region),
							},
							BlastPropagation: &sdp.BlastPropagation{
								// Changing the target group could affect the LB
								In: true,
								// The LB could also affect the target group
								Out: true,
							},
						})
					}
				}
			}
		}

		if action.RedirectConfig != nil {
			u := url.URL{}

			if action.RedirectConfig.Path != nil {
				u.Path = *action.RedirectConfig.Path
			}

			if action.RedirectConfig.Port != nil {
				u.Port()
			}

			if action.RedirectConfig.Host != nil {
				u.Host = *action.RedirectConfig.Host

				if action.RedirectConfig.Port != nil {
					u.Host = u.Host + fmt.Sprintf(":%v", *action.RedirectConfig.Port)
				}
			}

			if action.RedirectConfig.Protocol != nil {
				u.Scheme = *action.RedirectConfig.Protocol
			}

			if action.RedirectConfig.Query != nil {
				u.RawQuery = *action.RedirectConfig.Query
			}

			if u.Scheme == "http" || u.Scheme == "https" {
				requests = append(requests, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "http",
						Method: sdp.QueryMethod_GET,
						Query:  u.String(),
						Scope:  "global",
					},
					BlastPropagation: &sdp.BlastPropagation{
						// These are closely linked
						In:  true,
						Out: true,
					},
				})
			}
		}

		if action.TargetGroupArn != nil {
			if a, err := sources.ParseARN(*action.TargetGroupArn); err == nil {
				requests = append(requests, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "elbv2-target-group",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *action.TargetGroupArn,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					},
					BlastPropagation: &sdp.BlastPropagation{
						// These are closely linked
						In:  true,
						Out: true,
					},
				})
			}
		}
	}

	return requests
}
