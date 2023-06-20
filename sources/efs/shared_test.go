package efs

import (
	"github.com/overmindtech/aws-source/sources"
)

var TestRateLimit = sources.LimitBucket{
	MaxCapacity: 50,
	RefillRate:  20,
}
