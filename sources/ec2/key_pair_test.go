package ec2

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/sources"
)

func TestKeyPairInputMapperGet(t *testing.T) {
	input, err := KeyPairInputMapperGet("foo", "bar")

	if err != nil {
		t.Error(err)
	}

	if len(input.KeyNames) != 1 {
		t.Fatalf("expected 1 KeyPair ID, got %v", len(input.KeyNames))
	}

	if input.KeyNames[0] != "bar" {
		t.Errorf("expected KeyPair ID to be bar, got %v", input.KeyNames[0])
	}
}

func TestKeyPairInputMapperList(t *testing.T) {
	input, err := KeyPairInputMapperList("foo")

	if err != nil {
		t.Error(err)
	}

	if len(input.Filters) != 0 || len(input.KeyNames) != 0 {
		t.Errorf("non-empty input: %v", input)
	}
}

func TestKeyPairOutputMapper(t *testing.T) {
	output := &ec2.DescribeKeyPairsOutput{
		KeyPairs: []types.KeyPairInfo{
			{
				KeyPairId:      sources.PtrString("key-04d7068d3a33bf9b2"),
				KeyFingerprint: sources.PtrString("df:73:bb:86:a7:cd:9e:18:16:10:50:79:fa:3b:4f:c7:1d:32:cf:58"),
				KeyName:        sources.PtrString("dylan.ratcliffe"),
				KeyType:        types.KeyTypeRsa,
				Tags:           []types.Tag{},
				CreateTime:     sources.PtrTime(time.Now()),
				PublicKey:      sources.PtrString("PUB"),
			},
		},
	}

	items, err := KeyPairOutputMapper("foo", output)

	if err != nil {
		t.Fatal(err)
	}

	for _, item := range items {
		if err := item.Validate(); err != nil {
			t.Error(err)
		}
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %v", len(items))
	}

}

func TestNewKeyPairSource(t *testing.T) {
	config, account, _ := sources.GetAutoConfig(t)

	rateLimit := LimitBucket{
		MaxCapacity: 50,
		RefillRate:  10,
	}

	rateLimitCtx, rateLimitCancel := context.WithCancel(context.Background())
	defer rateLimitCancel()

	rateLimit.Start(rateLimitCtx)

	source := NewKeyPairSource(config, account, &rateLimit)

	test := sources.E2ETest{
		Source:  source,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
