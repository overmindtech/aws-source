package networkfirewall

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

type unifiedTLSInspectionConfiguration struct {
	Name                       string
	Properties                 *types.TLSInspectionConfigurationResponse
	TLSInspectionConfiguration *types.TLSInspectionConfiguration
}

func tlsInspectionConfigurationGetFunc(ctx context.Context, client networkFirewallClient, scope string, input *networkfirewall.DescribeTLSInspectionConfigurationInput) (*sdp.Item, error) {
	resp, err := client.DescribeTLSInspectionConfiguration(ctx, input)

	if err != nil {
		return nil, err
	}

	if resp == nil || resp.TLSInspectionConfiguration == nil || resp.TLSInspectionConfigurationResponse == nil ||
		resp.TLSInspectionConfigurationResponse.TLSInspectionConfigurationName == nil {

		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOTFOUND,
			ErrorString: "TLSInspectionConfiguration was nil",
			Scope:       scope,
		}
	}

	utic := unifiedTLSInspectionConfiguration{
		Name:                       *resp.TLSInspectionConfigurationResponse.TLSInspectionConfigurationName,
		Properties:                 resp.TLSInspectionConfigurationResponse,
		TLSInspectionConfiguration: resp.TLSInspectionConfiguration,
	}

	attributes, err := sources.ToAttributesCase(utic)

	if err != nil {
		return nil, err
	}

	tags := make(map[string]string)

	for _, tag := range resp.TLSInspectionConfigurationResponse.Tags {
		tags[*tag.Key] = *tag.Value
	}

	var health *sdp.Health

	switch resp.TLSInspectionConfigurationResponse.TLSInspectionConfigurationStatus {
	case types.ResourceStatusActive:
		health = sdp.Health_HEALTH_OK.Enum()
	case types.ResourceStatusDeleting:
		health = sdp.Health_HEALTH_PENDING.Enum()
	case types.ResourceStatusError:
		health = sdp.Health_HEALTH_ERROR.Enum()
	}

	item := sdp.Item{
		Type:            "network-firewall-tls-inspection-configuration",
		UniqueAttribute: "name",
		Attributes:      attributes,
		Scope:           scope,
		Tags:            tags,
		Health:          health,
	}

	if utic.Properties.CertificateAuthority != nil {
		if utic.Properties.CertificateAuthority.CertificateArn != nil {
			if a, err := sources.ParseARN(*utic.Properties.CertificateAuthority.CertificateArn); err == nil {
				//+overmind:link acm-pca-certificate-authority-certificate
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "acm-pca-certificate-authority-certificate",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *utic.Properties.CertificateAuthority.CertificateArn,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  true,
						Out: false,
					},
				})
			}
		}
	}

	for _, cert := range utic.Properties.Certificates {
		if cert.CertificateArn != nil {
			if a, err := sources.ParseARN(*cert.CertificateArn); err == nil {
				//+overmind:link acm-certificate
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "acm-certificate",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *cert.CertificateArn,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  true,
						Out: false,
					},
				})
			}
		}
	}

	item.LinkedItemQueries = append(item.LinkedItemQueries, encryptionConfigurationLink(utic.Properties.EncryptionConfiguration, scope))

	for _, config := range utic.TLSInspectionConfiguration.ServerCertificateConfigurations {
		if config.CertificateAuthorityArn != nil {
			if a, err := sources.ParseARN(*config.CertificateAuthorityArn); err == nil {
				//+overmind:link acm-pca-certificate-authority
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "acm-pca-certificate-authority",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *config.CertificateAuthorityArn,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  true,
						Out: false,
					},
				})
			}
		}

		for _, serverCert := range config.ServerCertificates {
			if serverCert.ResourceArn != nil {
				if a, err := sources.ParseARN(*serverCert.ResourceArn); err == nil {
					//+overmind:link acm-certificate
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "acm-certificate",
							Method: sdp.QueryMethod_SEARCH,
							Query:  *serverCert.ResourceArn,
							Scope:  sources.FormatScope(a.AccountID, a.Region),
						},
						BlastPropagation: &sdp.BlastPropagation{
							In:  true,
							Out: false,
						},
					})
				}
			}
		}
	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type network-firewall-tls-inspection-configuration
// +overmind:descriptiveType Network Firewall TLS Inspection Configuration
// +overmind:get Get a Network Firewall TLS Inspection Configuration by name
// +overmind:list List Network Firewall TLS Inspection Configurations
// +overmind:search Search for Network Firewall TLS Inspection Configurations by ARN
// +overmind:group AWS

func NewTLSInspectionConfigurationSource(client networkFirewallClient, accountID string, region string) *sources.AlwaysGetSource[*networkfirewall.ListTLSInspectionConfigurationsInput, *networkfirewall.ListTLSInspectionConfigurationsOutput, *networkfirewall.DescribeTLSInspectionConfigurationInput, *networkfirewall.DescribeTLSInspectionConfigurationOutput, networkFirewallClient, *networkfirewall.Options] {
	return &sources.AlwaysGetSource[*networkfirewall.ListTLSInspectionConfigurationsInput, *networkfirewall.ListTLSInspectionConfigurationsOutput, *networkfirewall.DescribeTLSInspectionConfigurationInput, *networkfirewall.DescribeTLSInspectionConfigurationOutput, networkFirewallClient, *networkfirewall.Options]{
		ItemType:  "network-firewall-tls-inspection-configuration",
		Client:    client,
		AccountID: accountID,
		Region:    region,
		ListInput: &networkfirewall.ListTLSInspectionConfigurationsInput{},
		GetInputMapper: func(scope, query string) *networkfirewall.DescribeTLSInspectionConfigurationInput {
			return &networkfirewall.DescribeTLSInspectionConfigurationInput{
				TLSInspectionConfigurationName: &query,
			}
		},
		SearchGetInputMapper: func(scope, query string) (*networkfirewall.DescribeTLSInspectionConfigurationInput, error) {
			return &networkfirewall.DescribeTLSInspectionConfigurationInput{
				TLSInspectionConfigurationArn: &query,
			}, nil
		},
		ListFuncPaginatorBuilder: func(client networkFirewallClient, input *networkfirewall.ListTLSInspectionConfigurationsInput) sources.Paginator[*networkfirewall.ListTLSInspectionConfigurationsOutput, *networkfirewall.Options] {
			return networkfirewall.NewListTLSInspectionConfigurationsPaginator(client, input)
		},
		ListFuncOutputMapper: func(output *networkfirewall.ListTLSInspectionConfigurationsOutput, input *networkfirewall.ListTLSInspectionConfigurationsInput) ([]*networkfirewall.DescribeTLSInspectionConfigurationInput, error) {
			var inputs []*networkfirewall.DescribeTLSInspectionConfigurationInput

			for _, rg := range output.TLSInspectionConfigurations {
				inputs = append(inputs, &networkfirewall.DescribeTLSInspectionConfigurationInput{
					TLSInspectionConfigurationArn: rg.Arn,
				})
			}
			return inputs, nil
		},
		GetFunc: func(ctx context.Context, client networkFirewallClient, scope string, input *networkfirewall.DescribeTLSInspectionConfigurationInput) (*sdp.Item, error) {
			return tlsInspectionConfigurationGetFunc(ctx, client, scope, input)
		},
	}
}
