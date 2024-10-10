package cloudfront

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/overmindtech/aws-source/adapters"
)

func TestKeyGroupItemMapper(t *testing.T) {
	group := types.KeyGroup{
		Id: adapters.PtrString("test-id"),
		KeyGroupConfig: &types.KeyGroupConfig{
			Items: []string{
				"some-identity",
			},
			Name:    adapters.PtrString("test-name"),
			Comment: adapters.PtrString("test-comment"),
		},
		LastModifiedTime: adapters.PtrTime(time.Now()),
	}

	item, err := KeyGroupItemMapper("", "test", &group)

	if err != nil {
		t.Fatal(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}
}

func TestNewKeyGroupAdapter(t *testing.T) {
	client, account, _ := GetAutoConfig(t)

	adapter := NewKeyGroupAdapter(client, account)

	test := adapters.E2ETest{
		Adapter: adapter,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
