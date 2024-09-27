package apigateway

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/overmindtech/aws-source/sources"
)

func deleteRestAPI(ctx context.Context, client *apigateway.Client, restAPIID string) error {
	_, err := client.DeleteRestApi(ctx, &apigateway.DeleteRestApiInput{
		RestApiId: sources.PtrString(restAPIID),
	})

	return err
}
