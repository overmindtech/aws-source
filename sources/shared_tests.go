package sources

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type Subnet struct {
	ID               *string
	CIDR             string
	AvailabilityZone string
}

type VPCConfig struct {
	// These are populated after creation
	ID                *string
	InternetGatewayId *string

	// CIDR to allocate *Required for create*
	CidrBlock string

	// Subnets to create
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

// Create Creates the VPC and subnets and registers cleanup actions for them
func (v *VPCConfig) Create(client *ec2.Client) error {
	var vpcOutput *ec2.CreateVpcOutput
	var err error

	vpcOutput, err = client.CreateVpc(
		context.Background(),
		&ec2.CreateVpcInput{
			CidrBlock: &v.CidrBlock,
			TagSpecifications: []types.TagSpecification{
				{
					ResourceType: types.ResourceTypeVpc,
					Tags:         TestTags,
				},
			},
		},
	)

	if err != nil {
		return err
	}

	v.ID = vpcOutput.Vpc.VpcId

	v.Cleanup(func() {
		var err error

		delete := func() error {
			_, err := client.DeleteVpc(
				context.Background(),
				&ec2.DeleteVpcInput{
					VpcId: v.ID,
				},
			)

			return err
		}

		retry(10, time.Second, delete)

		if err != nil {
			log.Println(err)
		}
	})

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

	v.InternetGatewayId = gatewayOutput.InternetGateway.InternetGatewayId

	v.Cleanup(func() {
		delete := func() error {
			_, err := client.DeleteInternetGateway(
				context.Background(),
				&ec2.DeleteInternetGatewayInput{
					InternetGatewayId: v.InternetGatewayId,
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
			InternetGatewayId: v.InternetGatewayId,
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
					InternetGatewayId: v.InternetGatewayId,
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

	for _, subnet := range v.Subnets {
		// Create subnets
		var subnetOutput *ec2.CreateSubnetOutput
		var err error

		subnetOutput, err = client.CreateSubnet(
			context.Background(),
			&ec2.CreateSubnetInput{
				VpcId:            v.ID,
				AvailabilityZone: &subnet.AvailabilityZone,
				CidrBlock:        &subnet.CIDR,
				TagSpecifications: []types.TagSpecification{
					{
						ResourceType: types.ResourceTypeSubnet,
						Tags:         TestTags,
					},
				},
			},
		)

		if err != nil {
			return err
		}

		subnet.ID = subnetOutput.Subnet.SubnetId

	}

	v.Cleanup(func() {
		for _, subnet := range v.Subnets {
			delete := func() error {
				_, err := client.DeleteSubnet(
					context.Background(),
					&ec2.DeleteSubnetInput{
						SubnetId: subnet.ID,
					},
				)

				return err
			}

			retry(10, time.Second, delete)

			if err != nil {
				log.Println(err)
			}
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
