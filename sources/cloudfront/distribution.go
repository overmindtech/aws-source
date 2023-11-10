package cloudfront

import (
	"context"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

var s3DnsRegex = regexp.MustCompile(`([^\.]+)\.s3\.([^\.]+)\.amazonaws\.com`)

func distributionGetFunc(ctx context.Context, client CloudFrontClient, scope string, input *cloudfront.GetDistributionInput) (*sdp.Item, error) {
	out, err := client.GetDistribution(ctx, input)

	if err != nil {
		return nil, err
	}

	d := out.Distribution

	if d == nil {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOTFOUND,
			ErrorString: "distribution was nil",
		}
	}

	var tags map[string]string

	// get tags
	tagsOut, err := client.ListTagsForResource(ctx, &cloudfront.ListTagsForResourceInput{
		Resource: d.ARN,
	})

	if err == nil {
		tags = tagsToMap(tagsOut.Tags)
	} else {
		tags = sources.HandleTagsError(ctx, err)
	}

	attributes, err := sources.ToAttributesCase(d)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "cloudfront-distribution",
		UniqueAttribute: "id",
		Attributes:      attributes,
		Scope:           scope,
		Tags:            tags,
	}

	if err != nil {
		return nil, err
	}

	if d.Status != nil {
		switch *d.Status {
		case "InProgress":
			item.Health = sdp.Health_HEALTH_PENDING.Enum()
		case "Deployed":
			item.Health = sdp.Health_HEALTH_OK.Enum()
		}
	}

	if d.DomainName != nil {
		// +overmind:link dns
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "dns",
				Method: sdp.QueryMethod_SEARCH,
				Query:  *d.DomainName,
				Scope:  "global",
			},
			BlastPropagation: &sdp.BlastPropagation{
				// DNS is always linked
				In:  true,
				Out: true,
			},
		})
	}

	if d.ActiveTrustedKeyGroups != nil {
		for _, keyGroup := range d.ActiveTrustedKeyGroups.Items {
			if keyGroup.KeyGroupId != nil {
				// +overmind:link cloudfront-key-group
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "cloudfront-key-group",
						Method: sdp.QueryMethod_GET,
						Query:  *keyGroup.KeyGroupId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// The distribution won't affect the key group
						Out: false,
						// The key group could affect the distribution
						In: true,
					},
				})
			}
		}
	}

	for _, record := range d.AliasICPRecordals {
		if record.CNAME != nil {
			// +overmind:link dns
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "dns",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *record.CNAME,
					Scope:  "global",
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Tightly linked
					In:  true,
					Out: true,
				},
			})
		}
	}

	if dc := d.DistributionConfig; dc != nil {
		if dc.Aliases != nil {
			for _, alias := range dc.Aliases.Items {
				// +overmind:link dns
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "dns",
						Method: sdp.QueryMethod_SEARCH,
						Query:  alias,
						Scope:  "global",
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Tightly linked
						In:  true,
						Out: true,
					},
				})
			}
		}

		if dc.ContinuousDeploymentPolicyId != nil && *dc.ContinuousDeploymentPolicyId != "" {
			// +overmind:link cloudfront-continuous-deployment-policy
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "cloudfront-continuous-deployment-policy",
					Method: sdp.QueryMethod_GET,
					Query:  *dc.ContinuousDeploymentPolicyId,
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					// These are tightly linked
					Out: true,
					In:  true,
				},
			})
		}

		if dc.CacheBehaviors != nil {
			for _, behavior := range dc.CacheBehaviors.Items {
				if behavior.CachePolicyId != nil {
					// +overmind:link cloudfront-cache-policy
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "cloudfront-cache-policy",
							Method: sdp.QueryMethod_GET,
							Query:  *behavior.CachePolicyId,
							Scope:  scope,
						},
						BlastPropagation: &sdp.BlastPropagation{
							// Changing the policy will affect the distribution
							In: true,
							// The distribution won't affect the policy
							Out: false,
						},
					})
				}

				if behavior.FieldLevelEncryptionId != nil {
					// +overmind:link cloudfront-field-level-encryption
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "cloudfront-field-level-encryption",
							Method: sdp.QueryMethod_GET,
							Query:  *behavior.FieldLevelEncryptionId,
							Scope:  scope,
						},
						BlastPropagation: &sdp.BlastPropagation{
							// Changing the encryption will affect the distribution
							In: true,
							// The distribution won't affect the encryption
							Out: false,
						},
					})
				}

				if behavior.OriginRequestPolicyId != nil {
					// +overmind:link cloudfront-origin-request-policy
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "cloudfront-origin-request-policy",
							Method: sdp.QueryMethod_GET,
							Query:  *behavior.OriginRequestPolicyId,
							Scope:  scope,
						},
						BlastPropagation: &sdp.BlastPropagation{
							// Changing the policy will affect the distribution
							In: true,
							// The distribution won't affect the policy
							Out: false,
						},
					})
				}

				if behavior.RealtimeLogConfigArn != nil {
					if arn, err := sources.ParseARN(*behavior.RealtimeLogConfigArn); err == nil {
						// +overmind:link cloudfront-realtime-log-config
						item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
							Query: &sdp.Query{
								Type:   "cloudfront-realtime-log-config",
								Method: sdp.QueryMethod_SEARCH,
								Query:  *behavior.RealtimeLogConfigArn,
								Scope:  sources.FormatScope(arn.AccountID, arn.Region),
							},
							BlastPropagation: &sdp.BlastPropagation{
								// Changing the config will affect the distribution
								In: true,
								// The distribution won't affect the config
								Out: false,
							},
						})
					}
				}

				if behavior.ResponseHeadersPolicyId != nil {
					// +overmind:link cloudfront-response-headers-policy
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "cloudfront-response-headers-policy",
							Method: sdp.QueryMethod_GET,
							Query:  *behavior.ResponseHeadersPolicyId,
							Scope:  scope,
						},
						BlastPropagation: &sdp.BlastPropagation{
							// Changing the policy will affect the distribution
							In: true,
							// The distribution won't affect the policy
							Out: false,
						},
					})
				}

				if behavior.TrustedKeyGroups != nil {
					for _, keyGroup := range behavior.TrustedKeyGroups.Items {
						// +overmind:link cloudfront-key-group
						item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
							Query: &sdp.Query{
								Type:   "cloudfront-key-group",
								Query:  keyGroup,
								Method: sdp.QueryMethod_GET,
								Scope:  scope,
							},
							BlastPropagation: &sdp.BlastPropagation{
								// Changing the key group will affect the distribution
								In: true,
								// The distribution won't affect the key group
								Out: false,
							},
						})
					}
				}

				if behavior.FunctionAssociations != nil {
					for _, function := range behavior.FunctionAssociations.Items {
						if function.FunctionARN != nil {
							if arn, err := sources.ParseARN(*function.FunctionARN); err == nil {
								// +overmind:link cloudfront-function
								item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
									Query: &sdp.Query{
										Type:   "cloudfront-function",
										Method: sdp.QueryMethod_SEARCH,
										Query:  *function.FunctionARN,
										Scope:  sources.FormatScope(arn.AccountID, arn.Region),
									},
									BlastPropagation: &sdp.BlastPropagation{
										// Changing the function could affect the distribution
										In: true,
										// The distribution could affect the function
										Out: true,
									},
								})
							}
						}
					}
				}

				if behavior.LambdaFunctionAssociations != nil {
					for _, function := range behavior.LambdaFunctionAssociations.Items {
						if arn, err := sources.ParseARN(*function.LambdaFunctionARN); err == nil {
							// +overmind:link lambda-function
							item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
								Query: &sdp.Query{
									Type:   "lambda-function",
									Method: sdp.QueryMethod_SEARCH,
									Query:  *function.LambdaFunctionARN,
									Scope:  sources.FormatScope(arn.AccountID, arn.Region),
								},
								BlastPropagation: &sdp.BlastPropagation{
									// Changing the function could affect the distribution
									In: true,
									// The distribution could affect the function
									Out: true,
								},
							})
						}
					}
				}
			}
		}

		if dc.Origins != nil {
			for _, origin := range dc.Origins.Items {
				if origin.DomainName != nil {
					// +overmind:link dns
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "dns",
							Method: sdp.QueryMethod_SEARCH,
							Query:  *origin.DomainName,
							Scope:  "global",
						},
						BlastPropagation: &sdp.BlastPropagation{
							// Tightly linked
							In:  true,
							Out: true,
						},
					})
				}

				if origin.OriginAccessControlId != nil {
					// +overmind:link cloudfront-origin-access-control
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "cloudfront-origin-access-control",
							Method: sdp.QueryMethod_GET,
							Query:  *origin.OriginAccessControlId,
							Scope:  scope,
						},
						BlastPropagation: &sdp.BlastPropagation{
							// Changing the access identity will affect the distribution
							In: true,
							// The distribution won't affect the access identity
							Out: false,
						},
					})
				}

				if origin.S3OriginConfig != nil {
					// If this is set then the origin is an S3 bucket, so we can
					// try to get the bucket name from the domain name
					if origin.DomainName != nil {
						matches := s3DnsRegex.FindStringSubmatch(*origin.DomainName)

						if len(matches) == 3 {
							// +overmind:link s3-bucket
							item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
								Query: &sdp.Query{
									Type:   "s3-bucket",
									Method: sdp.QueryMethod_GET,
									Query:  matches[1],
									Scope:  sources.FormatScope(scope, ""), // S3 buckets are global
								},
								BlastPropagation: &sdp.BlastPropagation{
									// Changing the bucket could affect the distribution
									In: true,
									// The distribution could affect the bucket
									Out: true,
								},
							})
						}
					}

					if origin.S3OriginConfig.OriginAccessIdentity != nil {
						// +overmind:link cloudfront-cloud-front-origin-access-identity
						item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
							Query: &sdp.Query{
								Type:   "cloudfront-cloud-front-origin-access-identity",
								Method: sdp.QueryMethod_GET,
								Query:  *origin.S3OriginConfig.OriginAccessIdentity,
								Scope:  scope,
							},
							BlastPropagation: &sdp.BlastPropagation{
								// Changing the access identity will affect the distribution
								In: true,
								// The distribution won't affect the access identity
								Out: false,
							},
						})
					}
				}
			}
		}

		if dc.DefaultCacheBehavior != nil {
			if dc.DefaultCacheBehavior.CachePolicyId != nil {
				// +overmind:link cloudfront-cache-policy
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "cloudfront-cache-policy",
						Method: sdp.QueryMethod_GET,
						Query:  *dc.DefaultCacheBehavior.CachePolicyId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changing the policy will affect the distribution
						In: true,
						// The distribution won't affect the policy
						Out: false,
					},
				})
			}

			if dc.DefaultCacheBehavior.FieldLevelEncryptionId != nil {
				// +overmind:link cloudfront-field-level-encryption
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "cloudfront-field-level-encryption",
						Method: sdp.QueryMethod_GET,
						Query:  *dc.DefaultCacheBehavior.FieldLevelEncryptionId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changing the encryption will affect the distribution
						In: true,
						// The distribution won't affect the encryption
						Out: false,
					},
				})
			}

			if dc.DefaultCacheBehavior.OriginRequestPolicyId != nil {
				// +overmind:link cloudfront-origin-request-policy
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "cloudfront-origin-request-policy",
						Method: sdp.QueryMethod_GET,
						Query:  *dc.DefaultCacheBehavior.OriginRequestPolicyId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changing the policy will affect the distribution
						In: true,
						// The distribution won't affect the policy
						Out: false,
					},
				})
			}

			if dc.DefaultCacheBehavior.RealtimeLogConfigArn != nil {
				if arn, err := sources.ParseARN(*dc.DefaultCacheBehavior.RealtimeLogConfigArn); err == nil {
					// +overmind:link cloudfront-realtime-log-config
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "cloudfront-realtime-log-config",
							Method: sdp.QueryMethod_GET,
							Query:  *dc.DefaultCacheBehavior.RealtimeLogConfigArn,
							Scope:  sources.FormatScope(arn.AccountID, arn.Region),
						},
						BlastPropagation: &sdp.BlastPropagation{
							// Changing the config will affect the distribution
							In: true,
							// The distribution won't affect the config
							Out: false,
						},
					})
				}
			}

			if dc.DefaultCacheBehavior.ResponseHeadersPolicyId != nil {
				// +overmind:link cloudfront-response-headers-policy
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "cloudfront-response-headers-policy",
						Method: sdp.QueryMethod_GET,
						Query:  *dc.DefaultCacheBehavior.ResponseHeadersPolicyId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changing the policy will affect the distribution
						In: true,
						// The distribution won't affect the policy
						Out: false,
					},
				})
			}

			if dc.DefaultCacheBehavior.TrustedKeyGroups != nil {
				for _, keyGroup := range dc.DefaultCacheBehavior.TrustedKeyGroups.Items {
					// +overmind:link cloudfront-key-group
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "cloudfront-key-group",
							Query:  keyGroup,
							Method: sdp.QueryMethod_GET,
							Scope:  scope,
						},
						BlastPropagation: &sdp.BlastPropagation{
							// Changing the key group will affect the distribution
							In: true,
							// The distribution won't affect the key group
							Out: false,
						},
					})
				}
			}

			if dc.DefaultCacheBehavior.FunctionAssociations != nil {
				for _, function := range dc.DefaultCacheBehavior.FunctionAssociations.Items {
					if function.FunctionARN != nil {
						if arn, err := sources.ParseARN(*function.FunctionARN); err == nil {
							// +overmind:link cloudfront-function
							item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
								Query: &sdp.Query{
									Type:   "cloudfront-function",
									Method: sdp.QueryMethod_SEARCH,
									Query:  *function.FunctionARN,
									Scope:  sources.FormatScope(arn.AccountID, arn.Region),
								},
								BlastPropagation: &sdp.BlastPropagation{
									// Changing the function could affect the distribution
									In: true,
									// The distribution could affect the function
									Out: true,
								},
							})
						}
					}
				}
			}

			if dc.DefaultCacheBehavior.LambdaFunctionAssociations != nil {
				for _, function := range dc.DefaultCacheBehavior.LambdaFunctionAssociations.Items {
					if arn, err := sources.ParseARN(*function.LambdaFunctionARN); err == nil {
						// +overmind:link lambda-function
						item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
							Query: &sdp.Query{
								Type:   "lambda-function",
								Method: sdp.QueryMethod_SEARCH,
								Query:  *function.LambdaFunctionARN,
								Scope:  sources.FormatScope(arn.AccountID, arn.Region),
							},
							BlastPropagation: &sdp.BlastPropagation{
								// Changing the function could affect the distribution
								In: true,
								// The distribution could affect the function
								Out: true,
							},
						})
					}
				}
			}
		}

		if dc.Logging != nil && dc.Logging.Bucket != nil {
			// +overmind:link dns
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "dns",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *dc.Logging.Bucket,
					Scope:  "global",
				},
				BlastPropagation: &sdp.BlastPropagation{
					// Tightly linked
					In:  true,
					Out: true,
				},
			})
		}

		if dc.ViewerCertificate != nil {
			if dc.ViewerCertificate.ACMCertificateArn != nil {
				if arn, err := sources.ParseARN(*dc.ViewerCertificate.ACMCertificateArn); err == nil {
					// +overmind:link acm-certificate
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "acm-certificate",
							Method: sdp.QueryMethod_SEARCH,
							Query:  *dc.ViewerCertificate.ACMCertificateArn,
							Scope:  sources.FormatScope(arn.AccountID, arn.Region),
						},
						BlastPropagation: &sdp.BlastPropagation{
							// Changing the certificate could affect the distribution
							In: true,
							// The distribution could not affect the certificate
							Out: false,
						},
					})
				}
			}
			if dc.ViewerCertificate.IAMCertificateId != nil {
				// +overmind:link iam-server-certificate
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "iam-server-certificate",
						Method: sdp.QueryMethod_GET,
						Query:  *dc.ViewerCertificate.IAMCertificateId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changing the certificate could affect the distribution
						In: true,
						// The distribution could not affect the certificate
						Out: false,
					},
				})
			}
		}

		if dc.WebACLId != nil {
			if arn, err := sources.ParseARN(*dc.WebACLId); err == nil {
				// +overmind:link wafv2-web-acl
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "wafv2-web-acl",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *dc.WebACLId,
						Scope:  sources.FormatScope(arn.AccountID, arn.Region),
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changing the ACL could affect the distribution
						In: true,
						// The distribution could not affect the ACL
						Out: false,
					},
				})
			} else {
				// Else assume it's a V1 ID
				// +overmind:link waf-web-acl
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "waf-web-acl",
						Method: sdp.QueryMethod_GET,
						Query:  *dc.WebACLId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changing the ACL could affect the distribution
						In: true,
						// The distribution could not affect the ACL
						Out: false,
					},
				})
			}
		}
	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type cloudfront-distribution
// +overmind:descriptiveType CloudFront Distribution
// +overmind:get
// +overmind:list
// +overmind:search
// +overmind:group AWS
// +overmind:terraform:queryMap aws_cloudfront_distribution.arn
// +overmind:terraform:method SEARCH

func NewDistributionSource(config aws.Config, accountID string) *sources.AlwaysGetSource[*cloudfront.ListDistributionsInput, *cloudfront.ListDistributionsOutput, *cloudfront.GetDistributionInput, *cloudfront.GetDistributionOutput, CloudFrontClient, *cloudfront.Options] {
	return &sources.AlwaysGetSource[*cloudfront.ListDistributionsInput, *cloudfront.ListDistributionsOutput, *cloudfront.GetDistributionInput, *cloudfront.GetDistributionOutput, CloudFrontClient, *cloudfront.Options]{
		ItemType:  "cloudfront-distribution",
		Client:    cloudfront.NewFromConfig(config),
		AccountID: accountID,
		Region:    "", // Cloudfront resources aren't tied to a region
		ListInput: &cloudfront.ListDistributionsInput{},
		ListFuncPaginatorBuilder: func(client CloudFrontClient, input *cloudfront.ListDistributionsInput) sources.Paginator[*cloudfront.ListDistributionsOutput, *cloudfront.Options] {
			return cloudfront.NewListDistributionsPaginator(client, input)
		},
		GetInputMapper: func(scope, query string) *cloudfront.GetDistributionInput {
			return &cloudfront.GetDistributionInput{
				Id: &query,
			}
		},
		ListFuncOutputMapper: func(output *cloudfront.ListDistributionsOutput, input *cloudfront.ListDistributionsInput) ([]*cloudfront.GetDistributionInput, error) {
			var inputs []*cloudfront.GetDistributionInput

			for _, distribution := range output.DistributionList.Items {
				inputs = append(inputs, &cloudfront.GetDistributionInput{
					Id: distribution.Id,
				})
			}

			return inputs, nil
		},
		GetFunc: distributionGetFunc,
	}
}
