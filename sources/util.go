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
	Region     string
	AccountID  string
	ResourceID string
}

// ParseARN Parses an ARN and tries to determine the resource ID from it. The
// logic is that the resource ID will be the last component when separated by
// slashes or colons: https://devopscube.com/aws-arn-guide/
func ParseARN(arnString string) (*ARN, error) {
	a, err := arn.Parse(arnString)

	if err != nil {
		return nil, err
	}

	fields := strings.FieldsFunc(a.Resource, func(r rune) bool {
		return r == '/' || r == ':'
	})

	return &ARN{
		Region:     a.Region,
		AccountID:  a.AccountID,
		ResourceID: fields[len(fields)-1],
	}, nil
}
