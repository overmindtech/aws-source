package ec2

import (
	"context"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources/integration"
)

const instanceSrc = "instance"

func setup(ctx context.Context, logger *slog.Logger, client *ec2.Client) error {
	// Create EC2 instance
	return createEC2Instance(ctx, logger, client, integration.TestID())
}
