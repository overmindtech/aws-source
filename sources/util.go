package sources

import "fmt"

// FormatScope Formats an account ID and region into the corresponding Overmind
// scope. This will be in the format {accountID}.{region}
func FormatScope(accountID, region string) string {
	return fmt.Sprintf("%v.%v", accountID, region)
}
