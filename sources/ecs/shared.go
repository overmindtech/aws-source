package ecs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

type ECSClient interface {
	DescribeClusters(ctx context.Context, params *ecs.DescribeClustersInput, optFns ...func(*ecs.Options)) (*ecs.DescribeClustersOutput, error)
	DescribeCapacityProviders(ctx context.Context, params *ecs.DescribeCapacityProvidersInput, optFns ...func(*ecs.Options)) (*ecs.DescribeCapacityProvidersOutput, error)

	ecs.ListClustersAPIClient
}
