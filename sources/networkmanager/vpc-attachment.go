package networkmanager

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func vpcAttachmentGetFunc(ctx context.Context, client *networkmanager.Client, scope, query string) (*types.VpcAttachment, error) {
	out, err := client.GetVpcAttachment(ctx, &networkmanager.GetVpcAttachmentInput{
		AttachmentId: &query,
	})
	if err != nil {
		return nil, err
	}

	return out.VpcAttachment, nil
}

//TODO: connect core-network here
//go:generate docgen ../../docs-data
// +overmind:type networkmanager-vpc-attachment
// +overmind:descriptiveType Networkmanager VPC Attachment
// +overmind:get Get a Networkmanager VPC Attachment by id
// +overmind:group AWS
// +overmind:terraform:queryMap aws_networkmanager_vpc_attachment.id

// TODO: connect coreNetwork here
func NewVPCAttachmentSource(config aws.Config, accountID string, limit *sources.LimitBucket) *sources.GetListSource[*types.VpcAttachment, *networkmanager.Client, *networkmanager.Options] {
	return &sources.GetListSource[*types.VpcAttachment, *networkmanager.Client, *networkmanager.Options]{
		Client:    networkmanager.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "networkmanager-vpc-attachment",
		GetFunc: func(ctx context.Context, client *networkmanager.Client, scope string, query string) (*types.VpcAttachment, error) {
			limit.Wait(ctx) // Wait for rate limiting
			return vpcAttachmentGetFunc(ctx, client, scope, query)
		},
		ListFunc: func(ctx context.Context, client *networkmanager.Client, scope string) ([]*types.VpcAttachment, error) {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_NOTFOUND,
				ErrorString: "list not supported for  networkmanager-vpc-attachment, use get",
			}
		},
	}
}
