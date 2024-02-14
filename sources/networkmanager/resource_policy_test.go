package networkmanager

import (
	"testing"
	"time"

	"github.com/overmindtech/aws-source/sources"
)

func TestNewResourcePolicy(t *testing.T) {
	config, account, _ := sources.GetAutoConfig(t)
	source := NewResourcePolicySource(config, account, &TestRateLimit)
	test := sources.E2ETest{
		Source:            source,
		Timeout:           30 * time.Second,
		SkipList:          true,
		SkipNotFoundCheck: true,
	}
	test.Run(t)
}
