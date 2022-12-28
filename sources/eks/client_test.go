package eks

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/eks"
)

type TestClient struct {
	ListClustersOutput    *eks.ListClustersOutput
	DescribeClusterOutput *eks.DescribeClusterOutput
	ListAddonsOutput      *eks.ListAddonsOutput
	DescribeAddonOutput   *eks.DescribeAddonOutput
}

func (t TestClient) ListClusters(context.Context, *eks.ListClustersInput, ...func(*eks.Options)) (*eks.ListClustersOutput, error) {
	return t.ListClustersOutput, nil
}

func (t TestClient) DescribeCluster(ctx context.Context, params *eks.DescribeClusterInput, optFns ...func(*eks.Options)) (*eks.DescribeClusterOutput, error) {
	return t.DescribeClusterOutput, nil
}

func (t TestClient) ListAddons(context.Context, *eks.ListAddonsInput, ...func(*eks.Options)) (*eks.ListAddonsOutput, error) {
	return t.ListAddonsOutput, nil
}

func (t TestClient) DescribeAddon(ctx context.Context, params *eks.DescribeAddonInput, optFns ...func(*eks.Options)) (*eks.DescribeAddonOutput, error) {
	return t.DescribeAddonOutput, nil
}
