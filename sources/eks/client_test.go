package eks

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/eks"
)

type TestClient struct {
	ListClustersOutput    *eks.ListClustersOutput
	DescribeClusterOutput *eks.DescribeClusterOutput
}

func (t TestClient) ListClusters(context.Context, *eks.ListClustersInput, ...func(*eks.Options)) (*eks.ListClustersOutput, error) {
	return t.ListClustersOutput, nil
}

func (t TestClient) DescribeCluster(ctx context.Context, params *eks.DescribeClusterInput, optFns ...func(*eks.Options)) (*eks.DescribeClusterOutput, error) {
	return t.DescribeClusterOutput, nil
}
