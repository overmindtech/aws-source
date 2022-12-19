package ec2

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestInstanceInputMapperGet(t *testing.T) {
	input, err := InstanceInputMapperGet("foo", "bar")

	if err != nil {
		t.Error(err)
	}

	if len(input.InstanceIds) != 1 {
		t.Fatalf("expected 1 instance ID, got %v", len(input.InstanceIds))
	}

	if input.InstanceIds[0] != "bar" {
		t.Errorf("expected instance ID to be bar, got %v", input.InstanceIds[0])
	}
}

func TestInstanceInputMapperList(t *testing.T) {
	input, err := InstanceInputMapperList("foo")

	if err != nil {
		t.Error(err)
	}

	if len(input.Filters) != 0 || len(input.InstanceIds) != 0 {
		t.Errorf("non-empty input: %v", input)
	}
}

func TestInstanceOutputMapper(t *testing.T) {
	output := &ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{
					{
						AmiLaunchIndex:  sources.PtrInt32(0),
						PublicIpAddress: sources.PtrString("43.5.36.7"),
						ImageId:         sources.PtrString("ami-04706e771f950937f"),
						InstanceId:      sources.PtrString("i-04c7b2794f7bc3d6a"),
						InstanceType:    types.InstanceTypeT2Micro,
						KeyName:         sources.PtrString("dylan.ratcliffe"),
						LaunchTime:      sources.PtrTime(time.Now()),
						Monitoring: &types.Monitoring{
							State: types.MonitoringStateDisabled,
						},
						Placement: &types.Placement{
							AvailabilityZone: sources.PtrString("eu-west-2c"),
							GroupName:        sources.PtrString(""),
							GroupId:          sources.PtrString("groupId"),
							Tenancy:          types.TenancyDefault,
						},
						PrivateDnsName:   sources.PtrString("ip-172-31-95-79.eu-west-2.compute.internal"),
						PrivateIpAddress: sources.PtrString("172.31.95.79"),
						ProductCodes:     []types.ProductCode{},
						PublicDnsName:    sources.PtrString(""),
						State: &types.InstanceState{
							Code: sources.PtrInt32(16),
							Name: types.InstanceStateNameRunning,
						},
						StateTransitionReason: sources.PtrString(""),
						SubnetId:              sources.PtrString("subnet-0450a637af9984235"),
						VpcId:                 sources.PtrString("vpc-0d7892e00e573e701"),
						Architecture:          types.ArchitectureValuesX8664,
						BlockDeviceMappings: []types.InstanceBlockDeviceMapping{
							{
								DeviceName: sources.PtrString("/dev/xvda"),
								Ebs: &types.EbsInstanceBlockDevice{
									AttachTime:          sources.PtrTime(time.Now()),
									DeleteOnTermination: sources.PtrBool(true),
									Status:              types.AttachmentStatusAttached,
									VolumeId:            sources.PtrString("vol-06c7211d9e79a355e"),
								},
							},
						},
						ClientToken:  sources.PtrString("eafad400-29e0-4b5c-a0fc-ef74c77659c4"),
						EbsOptimized: sources.PtrBool(false),
						EnaSupport:   sources.PtrBool(true),
						Hypervisor:   types.HypervisorTypeXen,
						NetworkInterfaces: []types.InstanceNetworkInterface{
							{
								Attachment: &types.InstanceNetworkInterfaceAttachment{
									AttachTime:          sources.PtrTime(time.Now()),
									AttachmentId:        sources.PtrString("eni-attach-02b19215d0dd9c7be"),
									DeleteOnTermination: sources.PtrBool(true),
									DeviceIndex:         sources.PtrInt32(0),
									Status:              types.AttachmentStatusAttached,
									NetworkCardIndex:    sources.PtrInt32(0),
								},
								Description: sources.PtrString(""),
								Groups: []types.GroupIdentifier{
									{
										GroupName: sources.PtrString("default"),
										GroupId:   sources.PtrString("sg-094e151c9fc5da181"),
									},
								},
								Ipv6Addresses:      []types.InstanceIpv6Address{},
								MacAddress:         sources.PtrString("02:8c:61:38:6f:c2"),
								NetworkInterfaceId: sources.PtrString("eni-09711a69e6d511358"),
								OwnerId:            sources.PtrString("052392120703"),
								PrivateDnsName:     sources.PtrString("ip-172-31-95-79.eu-west-2.compute.internal"),
								PrivateIpAddress:   sources.PtrString("172.31.95.79"),
								PrivateIpAddresses: []types.InstancePrivateIpAddress{
									{
										Primary:          sources.PtrBool(true),
										PrivateDnsName:   sources.PtrString("ip-172-31-95-79.eu-west-2.compute.internal"),
										PrivateIpAddress: sources.PtrString("172.31.95.79"),
									},
								},
								SourceDestCheck: sources.PtrBool(true),
								Status:          types.NetworkInterfaceStatusInUse,
								SubnetId:        sources.PtrString("subnet-0450a637af9984235"),
								VpcId:           sources.PtrString("vpc-0d7892e00e573e701"),
								InterfaceType:   sources.PtrString("interface"),
							},
						},
						RootDeviceName: sources.PtrString("/dev/xvda"),
						RootDeviceType: types.DeviceTypeEbs,
						SecurityGroups: []types.GroupIdentifier{
							{
								GroupName: sources.PtrString("default"),
								GroupId:   sources.PtrString("sg-094e151c9fc5da181"),
							},
						},
						SourceDestCheck: sources.PtrBool(true),
						Tags: []types.Tag{
							{
								Key:   sources.PtrString("Name"),
								Value: sources.PtrString("test"),
							},
						},
						VirtualizationType: types.VirtualizationTypeHvm,
						CpuOptions: &types.CpuOptions{
							CoreCount:      sources.PtrInt32(1),
							ThreadsPerCore: sources.PtrInt32(1),
						},
						CapacityReservationSpecification: &types.CapacityReservationSpecificationResponse{
							CapacityReservationPreference: types.CapacityReservationPreferenceOpen,
						},
						HibernationOptions: &types.HibernationOptions{
							Configured: sources.PtrBool(false),
						},
						MetadataOptions: &types.InstanceMetadataOptionsResponse{
							State:                   types.InstanceMetadataOptionsStateApplied,
							HttpTokens:              types.HttpTokensStateOptional,
							HttpPutResponseHopLimit: sources.PtrInt32(1),
							HttpEndpoint:            types.InstanceMetadataEndpointStateEnabled,
							HttpProtocolIpv6:        types.InstanceMetadataProtocolStateDisabled,
							InstanceMetadataTags:    types.InstanceMetadataTagsStateDisabled,
						},
						EnclaveOptions: &types.EnclaveOptions{
							Enabled: sources.PtrBool(false),
						},
						PlatformDetails:          sources.PtrString("Linux/UNIX"),
						UsageOperation:           sources.PtrString("RunInstances"),
						UsageOperationUpdateTime: sources.PtrTime(time.Now()),
						PrivateDnsNameOptions: &types.PrivateDnsNameOptionsResponse{
							HostnameType:                    types.HostnameTypeIpName,
							EnableResourceNameDnsARecord:    sources.PtrBool(true),
							EnableResourceNameDnsAAAARecord: sources.PtrBool(false),
						},
						MaintenanceOptions: &types.InstanceMaintenanceOptions{
							AutoRecovery: types.InstanceAutoRecoveryStateDefault,
						},
					},
				},
			},
		},
	}

	items, err := InstanceOutputMapper("foo", output)

	if err != nil {
		t.Fatal(err)
	}

	for _, item := range items {
		if err := item.Validate(); err != nil {
			t.Error(err)
		}
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %v", len(items))
	}

	item := items[0]

	// It doesn't really make sense to test anything other than the linked items
	// since the attributes are converted automatically
	tests := sources.ItemRequestTests{
		{
			ExpectedType:   "ec2-image",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "ami-04706e771f950937f",
			ExpectedScope:  item.Scope,
		},
		{
			ExpectedType:   "ip",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "172.31.95.79",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "ec2-subnet",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "subnet-0450a637af9984235",
			ExpectedScope:  item.Scope,
		},
		{
			ExpectedType:   "ec2-vpc",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "vpc-0d7892e00e573e701",
			ExpectedScope:  item.Scope,
		},
		{
			ExpectedType:   "ip",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "43.5.36.7",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "ec2-security-group",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "sg-094e151c9fc5da181",
			ExpectedScope:  item.Scope,
		},
		{
			ExpectedType:   "ec2-instance-status",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "i-04c7b2794f7bc3d6a",
			ExpectedScope:  item.Scope,
		},
		{
			ExpectedType:   "ec2-volume",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "vol-06c7211d9e79a355e",
			ExpectedScope:  item.Scope,
		},
		{
			ExpectedType:   "ec2-placement-group",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "groupId",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)

}
