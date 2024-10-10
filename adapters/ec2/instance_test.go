package ec2

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func TestInstanceInputMapperGet(t *testing.T) {
	input, err := instanceInputMapperGet("foo", "bar")

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
	input, err := instanceInputMapperList("foo")

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
						AmiLaunchIndex:  adapters.PtrInt32(0),
						PublicIpAddress: adapters.PtrString("43.5.36.7"),
						ImageId:         adapters.PtrString("ami-04706e771f950937f"),
						InstanceId:      adapters.PtrString("i-04c7b2794f7bc3d6a"),
						IamInstanceProfile: &types.IamInstanceProfile{
							Arn: adapters.PtrString("arn:aws:iam::052392120703:instance-profile/test"),
							Id:  adapters.PtrString("AIDAJQEAZVQ7Y2EYQ2Z6Q"),
						},
						BootMode:                types.BootModeValuesLegacyBios,
						CurrentInstanceBootMode: types.InstanceBootModeValuesLegacyBios,
						ElasticGpuAssociations: []types.ElasticGpuAssociation{
							{
								ElasticGpuAssociationId:    adapters.PtrString("ega-0a1b2c3d4e5f6g7h8"),
								ElasticGpuAssociationState: adapters.PtrString("associated"),
								ElasticGpuAssociationTime:  adapters.PtrString("now"),
								ElasticGpuId:               adapters.PtrString("egp-0a1b2c3d4e5f6g7h8"),
							},
						},
						CapacityReservationId: adapters.PtrString("cr-0a1b2c3d4e5f6g7h8"),
						InstanceType:          types.InstanceTypeT2Micro,
						ElasticInferenceAcceleratorAssociations: []types.ElasticInferenceAcceleratorAssociation{
							{
								ElasticInferenceAcceleratorArn:              adapters.PtrString("arn:aws:elastic-inference:us-east-1:052392120703:accelerator/eia-0a1b2c3d4e5f6g7h8"),
								ElasticInferenceAcceleratorAssociationId:    adapters.PtrString("eiaa-0a1b2c3d4e5f6g7h8"),
								ElasticInferenceAcceleratorAssociationState: adapters.PtrString("associated"),
								ElasticInferenceAcceleratorAssociationTime:  adapters.PtrTime(time.Now()),
							},
						},
						InstanceLifecycle: types.InstanceLifecycleTypeScheduled,
						Ipv6Address:       adapters.PtrString("2001:db8:3333:4444:5555:6666:7777:8888"),
						KeyName:           adapters.PtrString("dylan.ratcliffe"),
						KernelId:          adapters.PtrString("aki-0a1b2c3d4e5f6g7h8"),
						Licenses: []types.LicenseConfiguration{
							{
								LicenseConfigurationArn: adapters.PtrString("arn:aws:license-manager:us-east-1:052392120703:license-configuration:lic-0a1b2c3d4e5f6g7h8"),
							},
						},
						OutpostArn:            adapters.PtrString("arn:aws:outposts:us-east-1:052392120703:outpost/op-0a1b2c3d4e5f6g7h8"),
						Platform:              types.PlatformValuesWindows,
						RamdiskId:             adapters.PtrString("ari-0a1b2c3d4e5f6g7h8"),
						SpotInstanceRequestId: adapters.PtrString("sir-0a1b2c3d4e5f6g7h8"),
						SriovNetSupport:       adapters.PtrString("simple"),
						StateReason: &types.StateReason{
							Code:    adapters.PtrString("foo"),
							Message: adapters.PtrString("bar"),
						},
						TpmSupport: adapters.PtrString("foo"),
						LaunchTime: adapters.PtrTime(time.Now()),
						Monitoring: &types.Monitoring{
							State: types.MonitoringStateDisabled,
						},
						Placement: &types.Placement{
							AvailabilityZone: adapters.PtrString("eu-west-2c"), // link
							GroupName:        adapters.PtrString(""),
							GroupId:          adapters.PtrString("groupId"),
							Tenancy:          types.TenancyDefault,
						},
						PrivateDnsName:   adapters.PtrString("ip-172-31-95-79.eu-west-2.compute.internal"),
						PrivateIpAddress: adapters.PtrString("172.31.95.79"),
						ProductCodes:     []types.ProductCode{},
						PublicDnsName:    adapters.PtrString(""),
						State: &types.InstanceState{
							Code: adapters.PtrInt32(16),
							Name: types.InstanceStateNameRunning,
						},
						StateTransitionReason: adapters.PtrString(""),
						SubnetId:              adapters.PtrString("subnet-0450a637af9984235"),
						VpcId:                 adapters.PtrString("vpc-0d7892e00e573e701"),
						Architecture:          types.ArchitectureValuesX8664,
						BlockDeviceMappings: []types.InstanceBlockDeviceMapping{
							{
								DeviceName: adapters.PtrString("/dev/xvda"),
								Ebs: &types.EbsInstanceBlockDevice{
									AttachTime:          adapters.PtrTime(time.Now()),
									DeleteOnTermination: adapters.PtrBool(true),
									Status:              types.AttachmentStatusAttached,
									VolumeId:            adapters.PtrString("vol-06c7211d9e79a355e"),
								},
							},
						},
						ClientToken:  adapters.PtrString("eafad400-29e0-4b5c-a0fc-ef74c77659c4"),
						EbsOptimized: adapters.PtrBool(false),
						EnaSupport:   adapters.PtrBool(true),
						Hypervisor:   types.HypervisorTypeXen,
						NetworkInterfaces: []types.InstanceNetworkInterface{
							{
								Attachment: &types.InstanceNetworkInterfaceAttachment{
									AttachTime:          adapters.PtrTime(time.Now()),
									AttachmentId:        adapters.PtrString("eni-attach-02b19215d0dd9c7be"),
									DeleteOnTermination: adapters.PtrBool(true),
									DeviceIndex:         adapters.PtrInt32(0),
									Status:              types.AttachmentStatusAttached,
									NetworkCardIndex:    adapters.PtrInt32(0),
								},
								Description: adapters.PtrString(""),
								Groups: []types.GroupIdentifier{
									{
										GroupName: adapters.PtrString("default"),
										GroupId:   adapters.PtrString("sg-094e151c9fc5da181"),
									},
								},
								Ipv6Addresses:      []types.InstanceIpv6Address{},
								MacAddress:         adapters.PtrString("02:8c:61:38:6f:c2"),
								NetworkInterfaceId: adapters.PtrString("eni-09711a69e6d511358"),
								OwnerId:            adapters.PtrString("052392120703"),
								PrivateDnsName:     adapters.PtrString("ip-172-31-95-79.eu-west-2.compute.internal"),
								PrivateIpAddress:   adapters.PtrString("172.31.95.79"),
								PrivateIpAddresses: []types.InstancePrivateIpAddress{
									{
										Primary:          adapters.PtrBool(true),
										PrivateDnsName:   adapters.PtrString("ip-172-31-95-79.eu-west-2.compute.internal"),
										PrivateIpAddress: adapters.PtrString("172.31.95.79"),
									},
								},
								SourceDestCheck: adapters.PtrBool(true),
								Status:          types.NetworkInterfaceStatusInUse,
								SubnetId:        adapters.PtrString("subnet-0450a637af9984235"),
								VpcId:           adapters.PtrString("vpc-0d7892e00e573e701"),
								InterfaceType:   adapters.PtrString("interface"),
							},
						},
						RootDeviceName: adapters.PtrString("/dev/xvda"),
						RootDeviceType: types.DeviceTypeEbs,
						SecurityGroups: []types.GroupIdentifier{
							{
								GroupName: adapters.PtrString("default"),
								GroupId:   adapters.PtrString("sg-094e151c9fc5da181"),
							},
						},
						SourceDestCheck: adapters.PtrBool(true),
						Tags: []types.Tag{
							{
								Key:   adapters.PtrString("Name"),
								Value: adapters.PtrString("test"),
							},
						},
						VirtualizationType: types.VirtualizationTypeHvm,
						CpuOptions: &types.CpuOptions{
							CoreCount:      adapters.PtrInt32(1),
							ThreadsPerCore: adapters.PtrInt32(1),
						},
						CapacityReservationSpecification: &types.CapacityReservationSpecificationResponse{
							CapacityReservationPreference: types.CapacityReservationPreferenceOpen,
						},
						HibernationOptions: &types.HibernationOptions{
							Configured: adapters.PtrBool(false),
						},
						MetadataOptions: &types.InstanceMetadataOptionsResponse{
							State:                   types.InstanceMetadataOptionsStateApplied,
							HttpTokens:              types.HttpTokensStateOptional,
							HttpPutResponseHopLimit: adapters.PtrInt32(1),
							HttpEndpoint:            types.InstanceMetadataEndpointStateEnabled,
							HttpProtocolIpv6:        types.InstanceMetadataProtocolStateDisabled,
							InstanceMetadataTags:    types.InstanceMetadataTagsStateDisabled,
						},
						EnclaveOptions: &types.EnclaveOptions{
							Enabled: adapters.PtrBool(false),
						},
						PlatformDetails:          adapters.PtrString("Linux/UNIX"),
						UsageOperation:           adapters.PtrString("RunInstances"),
						UsageOperationUpdateTime: adapters.PtrTime(time.Now()),
						PrivateDnsNameOptions: &types.PrivateDnsNameOptionsResponse{
							HostnameType:                    types.HostnameTypeIpName,
							EnableResourceNameDnsARecord:    adapters.PtrBool(true),
							EnableResourceNameDnsAAAARecord: adapters.PtrBool(false),
						},
						MaintenanceOptions: &types.InstanceMaintenanceOptions{
							AutoRecovery: types.InstanceAutoRecoveryStateDefault,
						},
					},
				},
			},
		},
	}

	items, err := instanceOutputMapper(context.Background(), nil, "foo", nil, output)

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
	tests := adapters.QueryTests{
		{
			ExpectedType:   "ec2-image",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "ami-04706e771f950937f",
			ExpectedScope:  item.GetScope(),
		},
		{
			ExpectedType:   "ip",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "172.31.95.79",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "ec2-subnet",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "subnet-0450a637af9984235",
			ExpectedScope:  item.GetScope(),
		},
		{
			ExpectedType:   "ec2-vpc",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "vpc-0d7892e00e573e701",
			ExpectedScope:  item.GetScope(),
		},
		{
			ExpectedType:   "iam-instance-profile",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:iam::052392120703:instance-profile/test",
			ExpectedScope:  "052392120703",
		},
		{
			ExpectedType:   "ec2-capacity-reservation",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "cr-0a1b2c3d4e5f6g7h8",
			ExpectedScope:  item.GetScope(),
		},
		{
			ExpectedType:   "ec2-elastic-gpu",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "egp-0a1b2c3d4e5f6g7h8",
			ExpectedScope:  item.GetScope(),
		},
		{
			ExpectedType:   "elastic-inference-accelerator",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:elastic-inference:us-east-1:052392120703:accelerator/eia-0a1b2c3d4e5f6g7h8",
			ExpectedScope:  "052392120703.us-east-1",
		},
		{
			ExpectedType:   "ip",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "2001:db8:3333:4444:5555:6666:7777:8888",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "license-manager-license-configuration",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:license-manager:us-east-1:052392120703:license-configuration:lic-0a1b2c3d4e5f6g7h8",
			ExpectedScope:  "052392120703.us-east-1",
		},
		{
			ExpectedType:   "outposts-outpost",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:aws:outposts:us-east-1:052392120703:outpost/op-0a1b2c3d4e5f6g7h8",
			ExpectedScope:  "052392120703.us-east-1",
		},
		{
			ExpectedType:   "ec2-spot-instance-request",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "sir-0a1b2c3d4e5f6g7h8",
			ExpectedScope:  item.GetScope(),
		},
		{
			ExpectedType:   "ip",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "43.5.36.7",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "ec2-security-group",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "sg-094e151c9fc5da181",
			ExpectedScope:  item.GetScope(),
		},
		{
			ExpectedType:   "ec2-instance-status",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "i-04c7b2794f7bc3d6a",
			ExpectedScope:  item.GetScope(),
		},
		{
			ExpectedType:   "ec2-volume",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "vol-06c7211d9e79a355e",
			ExpectedScope:  item.GetScope(),
		},
		{
			ExpectedType:   "ec2-placement-group",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "groupId",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)

}

func TestNewInstanceAdapter(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	adapter := NewInstanceAdapter(client, account, region)

	test := adapters.E2ETest{
		Adapter: adapter,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
