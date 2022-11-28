package securitygroup

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/overmindtech/aws-source/sources"
)

// Shared variables that are populated before tests are run. These can be used
// to that each doesn't need to load config each time
var TestAWSConfig aws.Config
var TestAccountID string
var TestContext string
var TestVPC = sources.VPCConfig{
	CidrBlock: "10.174.145.0/24",
	Subnets: []*sources.Subnet{
		{
			CIDR:             "10.174.145.0/28",
			AvailabilityZone: "eu-west-2a",
		},
		{
			CIDR:             "10.174.145.16/28",
			AvailabilityZone: "eu-west-2b",
		},
		{
			CIDR:             "10.174.145.32/28",
			AvailabilityZone: "eu-west-2c",
		},
	},
}

func TestMain(m *testing.M) {
	var err error

	TestAWSConfig, err = config.LoadDefaultConfig(context.TODO())

	if err != nil {
		log.Fatalf("Config load failed: %v", err)
		log.Println("Tests will be skipped as AWS config could not be loaded")

		os.Exit(1)
	}

	// Override region since the tests require this at the moment
	TestAWSConfig.Region = "eu-west-2"

	ec2Client := ec2.NewFromConfig(TestAWSConfig)

	err = TestVPC.Create(ec2Client)

	if err != nil {
		log.Println(err)
	}

	stsClient := sts.NewFromConfig(TestAWSConfig)

	var callerID *sts.GetCallerIdentityOutput

	callerID, err = stsClient.GetCallerIdentity(
		context.Background(),
		&sts.GetCallerIdentityInput{},
	)

	if err != nil {
		log.Println(err)
	}

	TestAccountID = *callerID.Account

	TestContext = fmt.Sprintf("%v.%v", TestAccountID, TestAWSConfig.Region)

	exitVal := m.Run()

	TestVPC.RunCleanup()
	os.Exit(exitVal)
}
