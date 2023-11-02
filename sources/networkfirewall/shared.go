package networkfirewall

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
)

type networkFirewallClient interface {
	DescribeFirewall(ctx context.Context, params *networkfirewall.DescribeFirewallInput, optFns ...func(*networkfirewall.Options)) (*networkfirewall.DescribeFirewallOutput, error)
	ListFirewalls(context.Context, *networkfirewall.ListFirewallsInput, ...func(*networkfirewall.Options)) (*networkfirewall.ListFirewallsOutput, error)
}
