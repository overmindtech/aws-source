package sources

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/sdp-go"
)

type Subnet struct {
	ID               *string
	CIDR             string
	AvailabilityZone string
}

type VPCConfig struct {
	// These are populated after Fetching
	ID *string

	// Subnets in this VPC
	Subnets []*Subnet

	cleanupFunctions []func()
}

var purposeKey = "Purpose"
var nameKey = "Name"
var tagValue = "automated-testing-" + time.Now().Format("2006-01-02T15:04:05.000Z")
var TestTags = []types.Tag{
	{
		Key:   &purposeKey,
		Value: &tagValue,
	},
	{
		Key:   &nameKey,
		Value: &tagValue,
	},
}

func (v *VPCConfig) Cleanup(f func()) {
	v.cleanupFunctions = append(v.cleanupFunctions, f)
}

func (v *VPCConfig) RunCleanup() {
	for len(v.cleanupFunctions) > 0 {
		n := len(v.cleanupFunctions) - 1 // Top element

		v.cleanupFunctions[n]()

		v.cleanupFunctions = v.cleanupFunctions[:n] // Pop
	}
}

// Fetch Fetches the VPC and subnets and registers cleanup actions for them
func (v *VPCConfig) Fetch(client *ec2.Client) error {
	// manually configured VPC in eu-west-2
	vpcid := "vpc-061f0bb58acec88ad"
	v.ID = &vpcid // vpcOutput.Vpc.VpcId
	filterName := "vpc-id"
	subnetOutput, err := client.DescribeSubnets(
		context.Background(),
		&ec2.DescribeSubnetsInput{
			Filters: []types.Filter{
				{
					Name:   &filterName,
					Values: []string{vpcid},
				},
			},
		},
	)

	if err != nil {
		return err
	}

	for _, subnet := range subnetOutput.Subnets {
		v.Subnets = append(v.Subnets, &Subnet{
			ID:               subnet.SubnetId,
			CIDR:             *subnet.CidrBlock,
			AvailabilityZone: *subnet.AvailabilityZone,
		})
	}

	return nil
}

// CreateGateway Creates a new internet gateway for the duration of the test to save 40$ per month vs running it 24/7
func (v *VPCConfig) CreateGateway(client *ec2.Client) error {
	var err error

	// Create internet gateway and assign to VPC
	var gatewayOutput *ec2.CreateInternetGatewayOutput

	gatewayOutput, err = client.CreateInternetGateway(
		context.Background(),
		&ec2.CreateInternetGatewayInput{
			TagSpecifications: []types.TagSpecification{
				{
					ResourceType: types.ResourceTypeInternetGateway,
					Tags:         TestTags,
				},
			},
		},
	)

	if err != nil {
		return err
	}

	internetGatewayId := gatewayOutput.InternetGateway.InternetGatewayId

	v.Cleanup(func() {
		delete := func() error {
			_, err := client.DeleteInternetGateway(
				context.Background(),
				&ec2.DeleteInternetGatewayInput{
					InternetGatewayId: internetGatewayId,
				},
			)

			return err
		}

		err := retry(10, time.Second, delete)

		if err != nil {
			log.Println(err)
		}
	})

	_, err = client.AttachInternetGateway(
		context.Background(),
		&ec2.AttachInternetGatewayInput{
			InternetGatewayId: internetGatewayId,
			VpcId:             v.ID,
		},
	)

	if err != nil {
		return err
	}

	v.Cleanup(func() {
		delete := func() error {
			_, err := client.DetachInternetGateway(
				context.Background(),
				&ec2.DetachInternetGatewayInput{
					InternetGatewayId: internetGatewayId,
					VpcId:             v.ID,
				},
			)

			return err
		}

		err := retry(10, time.Second, delete)

		if err != nil {
			log.Println(err)
		}
	})
	return nil
}

func retry(attempts int, sleep time.Duration, f func() error) (err error) {
	for i := 0; i < attempts; i++ {
		if i > 0 {
			time.Sleep(sleep)
			sleep *= 2
		}
		err = f()
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}

// CheckItemRequest Checks that an item request matches the expected params
func CheckItemRequest(t *testing.T, item *sdp.ItemRequest, itemName string, expectedType string, expectedQuery string, expectedScope string) {
	if item.Type != expectedType {
		t.Errorf("%s.Type '%v' != '%v'", itemName, item.Type, expectedType)
	}
	if item.Method != sdp.RequestMethod_GET {
		t.Errorf("%s.Method '%v' != '%v'", itemName, item.Method, sdp.RequestMethod_GET)
	}
	if item.Query != expectedQuery {
		t.Errorf("%s.Query '%v' != '%v'", itemName, item.Query, expectedQuery)
	}
	if item.Scope != expectedScope {
		t.Errorf("%s.Scope '%v' != '%v'", itemName, item.Scope, expectedScope)
	}
}
