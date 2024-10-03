package cloudfront

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/overmindtech/aws-source/sources"
)

func TestKeyGroupItemMapper(t *testing.T) {
	group := types.KeyGroup{
		Id: sources.PtrString("test-id"),
		KeyGroupConfig: &types.KeyGroupConfig{
			Items: []string{
				"some-identity",
			},
			Name:    sources.PtrString("test-name"),
			Comment: sources.PtrString("test-comment"),
		},
		LastModifiedTime: sources.PtrTime(time.Now()),
	}

	item, err := KeyGroupItemMapper("", "test", &group)

	if err != nil {
		t.Fatal(err)
	}

	if err = item.Validate(); err != nil {
		t.Error(err)
	}
}

func TestNewKeyGroupSource(t *testing.T) {
	client, account, _ := GetAutoConfig(t)

	source := NewKeyGroupSource(client, account)

	test := sources.E2ETest{
		Adapter: source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
