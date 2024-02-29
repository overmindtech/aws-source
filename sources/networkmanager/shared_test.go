package networkmanager

import (
	"github.com/overmindtech/aws-source/sources"
)

type TestClient struct{}

var TestRateLimit = sources.LimitBucket{
	MaxCapacity: 50,
	RefillRate:  20,
}
