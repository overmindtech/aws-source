package networkmanager

import (
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestVPCAttachmentItemMapper(t *testing.T) {
	input := types.VpcAttachment{
		Attachment: &types.Attachment{
			AttachmentId:  sources.PtrString("attachment1"),
			CoreNetworkId: sources.PtrString("corenetwork1"),
		},
	}
	scope := "123456789012.eu-west-2"
	item, err := vpcAttachmentItemMapper(scope, &input)

	if err != nil {
		t.Error(err)
	}
	if err := item.Validate(); err != nil {
		t.Error(err)
	}

	// Ensure unique attribute
	require.NotNil(t, item.Attributes)
	uniqueAttr, err := item.Attributes.Get("attachmentId")
	require.Nil(t, err)
	require.Equal(t, "attachment1", uniqueAttr.(string))

	tests := sources.QueryTests{
		{
			ExpectedType:   "networkmanager-core-network",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "corenetwork1",
			ExpectedScope:  scope,
		},
	}

	tests.Execute(t, item)
}
