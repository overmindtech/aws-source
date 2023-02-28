package elbv2

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/url"

	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func ListenerOutputMapper(scope string, _ *elbv2.DescribeListenersInput, output *elbv2.DescribeListenersOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, listener := range output.Listeners {
		// Redact the client secret and replace with the first 12 characters of
		// the SHA1 hash so that we can at least tell if it has changed
		for _, action := range listener.DefaultActions {
			if action.AuthenticateOidcConfig != nil {
				if action.AuthenticateOidcConfig.ClientSecret != nil {
					h := sha1.New()
					h.Write([]byte(*action.AuthenticateOidcConfig.ClientSecret))
					sha := base64.URLEncoding.EncodeToString(h.Sum(nil))

					if len(sha) > 12 {
						action.AuthenticateOidcConfig.ClientSecret = sources.PtrString(fmt.Sprintf("REDACTED (Version: %v)", sha[:11]))
					} else {
						action.AuthenticateOidcConfig.ClientSecret = sources.PtrString("[REDACTED]")
					}
				}
			}
		}

		attrs, err := sources.ToAttributesCase(listener)

		if err != nil {
			return nil, err
		}

		item := sdp.Item{
			Type:            "elbv2-listener",
			UniqueAttribute: "listenerArn",
			Attributes:      attrs,
			Scope:           scope,
		}

		if listener.LoadBalancerArn != nil {
			if a, err := sources.ParseARN(*listener.LoadBalancerArn); err == nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "elbv2-load-balancer",
					Method: sdp.RequestMethod_SEARCH,
					Query:  *listener.LoadBalancerArn,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				})
			}
		}

		for _, cert := range listener.Certificates {
			if cert.CertificateArn != nil {
				if a, err := sources.ParseARN(*cert.CertificateArn); err == nil {
					item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
						Type:   "acm-certificate",
						Method: sdp.RequestMethod_SEARCH,
						Query:  *cert.CertificateArn,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					})
				}
			}
		}

		for _, action := range listener.DefaultActions {
			if action.AuthenticateCognitoConfig != nil {
				if action.AuthenticateCognitoConfig.UserPoolArn != nil {
					if a, err := sources.ParseARN(*action.AuthenticateCognitoConfig.UserPoolArn); err == nil {
						item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
							Type:   "cognito-idp-user-pool",
							Method: sdp.RequestMethod_SEARCH,
							Query:  *action.AuthenticateCognitoConfig.UserPoolArn,
							Scope:  sources.FormatScope(a.AccountID, a.Region),
						})
					}
				}
			}

			if action.AuthenticateOidcConfig != nil {
				if action.AuthenticateOidcConfig.AuthorizationEndpoint != nil {
					item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
						Type:   "http",
						Method: sdp.RequestMethod_GET,
						Query:  *action.AuthenticateOidcConfig.AuthorizationEndpoint,
						Scope:  "global",
					})
				}

				if action.AuthenticateOidcConfig.TokenEndpoint != nil {
					item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
						Type:   "http",
						Method: sdp.RequestMethod_GET,
						Query:  *action.AuthenticateOidcConfig.TokenEndpoint,
						Scope:  "global",
					})
				}

				if action.AuthenticateOidcConfig.UserInfoEndpoint != nil {
					item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
						Type:   "http",
						Method: sdp.RequestMethod_GET,
						Query:  *action.AuthenticateOidcConfig.UserInfoEndpoint,
						Scope:  "global",
					})
				}

				if action.ForwardConfig != nil {
					for _, tg := range action.ForwardConfig.TargetGroups {
						if tg.TargetGroupArn != nil {
							if a, err := sources.ParseARN(*tg.TargetGroupArn); err == nil {
								item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
									Type:   "elbv2-target-group",
									Method: sdp.RequestMethod_SEARCH,
									Query:  *tg.TargetGroupArn,
									Scope:  sources.FormatScope(a.AccountID, a.Region),
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
						item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
							Type:   "http",
							Method: sdp.RequestMethod_GET,
							Query:  u.String(),
							Scope:  "global",
						})
					}
				}

				if action.TargetGroupArn != nil {
					if a, err := sources.ParseARN(*action.TargetGroupArn); err == nil {
						item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
							Type:   "elbv2-target-group",
							Method: sdp.RequestMethod_SEARCH,
							Query:  *action.TargetGroupArn,
							Scope:  sources.FormatScope(a.AccountID, a.Region),
						})
					}
				}
			}
		}

		items = append(items, &item)
	}

	return items, nil
}

func NewListenerSource(config aws.Config, accountID string) *sources.DescribeOnlySource[*elbv2.DescribeListenersInput, *elbv2.DescribeListenersOutput, *elbv2.Client, *elbv2.Options] {
	return &sources.DescribeOnlySource[*elbv2.DescribeListenersInput, *elbv2.DescribeListenersOutput, *elbv2.Client, *elbv2.Options]{
		Config:    config,
		Client:    elbv2.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "elbv2-listener",
		DescribeFunc: func(ctx context.Context, client *elbv2.Client, input *elbv2.DescribeListenersInput) (*elbv2.DescribeListenersOutput, error) {
			return client.DescribeListeners(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*elbv2.DescribeListenersInput, error) {
			return &elbv2.DescribeListenersInput{
				ListenerArns: []string{query},
			}, nil
		},
		InputMapperList: func(scope string) (*elbv2.DescribeListenersInput, error) {
			return nil, &sdp.ItemRequestError{
				ErrorType:   sdp.ItemRequestError_NOTFOUND,
				ErrorString: "list not supported for elbv2-listener, use search",
			}
		},
		PaginatorBuilder: func(client *elbv2.Client, params *elbv2.DescribeListenersInput) sources.Paginator[*elbv2.DescribeListenersOutput, *elbv2.Options] {
			return elbv2.NewDescribeListenersPaginator(client, params)
		},
		OutputMapper: ListenerOutputMapper,
	}
}
