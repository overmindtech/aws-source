package sources

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
)

// FormatScope Formats an account ID and region into the corresponding Overmind
// scope. This will be in the format {accountID}.{region}
func FormatScope(accountID, region string) string {
	if region == "" {
		return accountID
	}

	return fmt.Sprintf("%v.%v", accountID, region)
}

// A parsed representation of the parts of the ARN that Overmind needs to care
// about
type ARN struct {
	// The region that the resource is in e.g. eu-west-1
	Region string
	// The account ID e.g. 052392120704
	AccountID string
	// The type and name of the resources, this everything after the account
	// including a version if relevant e.g.
	// task-definition/ecs-template-ecs-demo-app:1
	Resource string
	// The ID of the resource, this is everything after the type and might also
	// include a version or other components depending on the service e.g.
	// ecs-template-ecs-demo-app:1 would be the ResourceID for
	// "arn:aws:ecs:eu-west-1:052392120703:task-definition/ecs-template-ecs-demo-app:1"
	ResourceID string
	// The name of the AWS service that the ARN relates to e.g. ecs
	Service string
}

// ParseARN Parses an ARN and tries to determine the resource ID from it. The
// logic is that the resource ID will be the last component when separated by
// slashes or colons: https://devopscube.com/aws-arn-guide/
func ParseARN(arnString string) (*ARN, error) {
	a, err := arn.Parse(arnString)

	if err != nil {
		return nil, err
	}

	// Find the first separator
	separatorLocation := strings.IndexFunc(a.Resource, func(r rune) bool {
		return r == '/' || r == ':'
	})

	// Remove the first field since this is the type, then keen the rest
	resourceID := a.Resource[separatorLocation+1:]

	return &ARN{
		Region:     a.Region,
		AccountID:  a.AccountID,
		Resource:   a.Resource,
		ResourceID: resourceID,
		Service:    a.Service,
	}, nil
}
