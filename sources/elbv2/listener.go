package elbv2

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"

	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func listenerOutputMapper(_ context.Context, _ *elbv2.Client, scope string, _ *elbv2.DescribeListenersInput, output *elbv2.DescribeListenersOutput) ([]*sdp.Item, error) {
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
				// +overmind:link elbv2-load-balancer
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "elbv2-load-balancer",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *listener.LoadBalancerArn,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Load balancers and their listeners are tightly coupled
						In:  true,
						Out: true,
					},
				})
			}
		}

		for _, cert := range listener.Certificates {
			if cert.CertificateArn != nil {
				if a, err := sources.ParseARN(*cert.CertificateArn); err == nil {
					// +overmind:link acm-certificate
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "acm-certificate",
							Method: sdp.QueryMethod_SEARCH,
							Query:  *cert.CertificateArn,
							Scope:  sources.FormatScope(a.AccountID, a.Region),
						},
						BlastPropagation: &sdp.BlastPropagation{
							// Changing the cert will affect the LB
							In: true,
							// The LB won't affect the cert
							Out: false,
						},
					})
				}
			}
		}

		var requests []*sdp.LinkedItemQuery

		for _, action := range listener.DefaultActions {
			// These types can be returned by `ActionToRequests()`
			// +overmind:link cognito-idp-user-pool
			// +overmind:link http
			// +overmind:link elbv2-target-group

			requests = ActionToRequests(action)
			item.LinkedItemQueries = append(item.LinkedItemQueries, requests...)
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type elbv2-listener
// +overmind:descriptiveType ELB Listener
// +overmind:get Get a listener by ARN
// +overmind:search Search for listeners by load balancer ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_alb_listener.arn
// +overmind:terraform:queryMap aws_lb_listener.arn
// +overmind:terraform:method SEARCH

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
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_NOTFOUND,
				ErrorString: "list not supported for elbv2-listener, use search",
			}
		},
		InputMapperSearch: func(ctx context.Context, client *elbv2.Client, scope, query string) (*elbv2.DescribeListenersInput, error) {
			// Search by LB ARN
			return &elbv2.DescribeListenersInput{
				LoadBalancerArn: &query,
			}, nil
		},
		PaginatorBuilder: func(client *elbv2.Client, params *elbv2.DescribeListenersInput) sources.Paginator[*elbv2.DescribeListenersOutput, *elbv2.Options] {
			return elbv2.NewDescribeListenersPaginator(client, params)
		},
		OutputMapper: listenerOutputMapper,
	}
}
