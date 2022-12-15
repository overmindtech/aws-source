package availabilityzone

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/sources"
)

func TestAvailabilityZonesMapping(t *testing.T) {
	t.Parallel()

	t.Run("empty", func(t *testing.T) {
		name := "eu-west-2"
		sg := types.AvailabilityZone{
			ZoneName: &name,
		}

		item, err := mapAvailabilityZoneToItem(&sg, "foo.bar")
		if err != nil {
			t.Fatal(err)
		}
		if item == nil {
			t.Fatal("item is nil")
		}
		if item.Attributes == nil || item.Attributes.AttrStruct.Fields["zoneName"].GetStringValue() != name {
			t.Errorf("unexpected item: %v", item)
		}
		if item.UniqueAttributeValue() != name {
			t.Errorf("expected UAV to be %v, got %v", name, item.UniqueAttributeValue())
		}
	})
	t.Run("Fully populated", func(t *testing.T) {
		message := "everything is fine"
		regionName := "eu-west-2"
		zoneName := "eu-west-2a"
		zoneId := "euw2-az2"
		groupName := "eu-west-2"
		networkBorderGroup := "eu-west-2"
		zoneType := "availability-zone"
		sg := types.AvailabilityZone{
			State:       types.AvailabilityZoneStateAvailable,
			OptInStatus: types.AvailabilityZoneOptInStatusOptInNotRequired,
			Messages: []types.AvailabilityZoneMessage{
				{
					Message: &message,
				},
			},
			RegionName:         &regionName,
			ZoneName:           &zoneName,
			ZoneId:             &zoneId,
			GroupName:          &groupName,
			NetworkBorderGroup: &networkBorderGroup,
			ZoneType:           &zoneType,
		}

		item, err := mapAvailabilityZoneToItem(&sg, "foo.bar")
		if err != nil {
			t.Fatal(err)
		}
		if item == nil {
			t.Fatal("item is nil")
		}
		if len(item.LinkedItemRequests) != 1 {
			t.Fatalf("unexpected LinkedItemRequests: %v", item)
		}
		sources.CheckItemRequest(t, item.LinkedItemRequests[0], "region", "ec2-region", "eu-west-2", item.Scope)
	})
}

func TestGet(t *testing.T) {
	t.Parallel()
	t.Run("empty (scope mismatch)", func(t *testing.T) {
		src := AvailabilityZoneSource{}

		items, err := src.Get(context.Background(), "foo.bar", "query")
		if items != nil {
			t.Fatalf("unexpected items: %v", items)
		}
		if err == nil {
			t.Fatalf("expected err, got nil")
		}
		if !strings.HasPrefix(err.Error(), "requested scope foo.bar does not match source scope .") {
			t.Errorf("expected 'requested scope foo.bar does not match source scope .', got '%v'", err.Error())
		}
	})
}

type fakeClient struct{}

func (m fakeClient) DescribeAvailabilityZones(ctx context.Context, params *ec2.DescribeAvailabilityZonesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeAvailabilityZonesOutput, error) {
	euWest2aMessage := "everything is fine"
	euWest2aRegionName := "eu-west-2"
	euWest2aZoneName := "eu-west-2a"
	euWest2aZoneId := "euw2-az2"
	euWest2aGroupName := "eu-west-2"
	euWest2aNetworkBorderGroup := "eu-west-2"
	euWest2aZoneType := "availability-zone"

	euWest2a := types.AvailabilityZone{
		State:       types.AvailabilityZoneStateAvailable,
		OptInStatus: types.AvailabilityZoneOptInStatusOptInNotRequired,
		Messages: []types.AvailabilityZoneMessage{
			{
				Message: &euWest2aMessage,
			},
		},
		RegionName:         &euWest2aRegionName,
		ZoneName:           &euWest2aZoneName,
		ZoneId:             &euWest2aZoneId,
		GroupName:          &euWest2aGroupName,
		NetworkBorderGroup: &euWest2aNetworkBorderGroup,
		ZoneType:           &euWest2aZoneType,
	}

	euWest2bRegionName := "eu-west-2"
	euWest2bZoneName := "eu-west-2b"
	euWest2bZoneId := "euw2-az3"
	euWest2bGroupName := "eu-west-2"
	euWest2bNetworkBorderGroup := "eu-west-2"
	euWest2bZoneType := "availability-zone"

	euWest2b := types.AvailabilityZone{
		State:              types.AvailabilityZoneStateAvailable,
		OptInStatus:        types.AvailabilityZoneOptInStatusOptInNotRequired,
		Messages:           []types.AvailabilityZoneMessage{},
		RegionName:         &euWest2bRegionName,
		ZoneName:           &euWest2bZoneName,
		ZoneId:             &euWest2bZoneId,
		GroupName:          &euWest2bGroupName,
		NetworkBorderGroup: &euWest2bNetworkBorderGroup,
		ZoneType:           &euWest2bZoneType,
	}

	testZones := []types.AvailabilityZone{
		euWest2a,
		euWest2b,
	}

	if len(params.ZoneNames) == 0 {
		return &ec2.DescribeAvailabilityZonesOutput{
			AvailabilityZones: testZones,
		}, nil
	}

	results := make([]types.AvailabilityZone, 0)

	for _, nameQuery := range params.ZoneNames {
		for _, zone := range testZones {
			if *zone.ZoneName == nameQuery {
				results = append(results, zone)
			}
		}
	}

	return &ec2.DescribeAvailabilityZonesOutput{
		AvailabilityZones: results,
	}, nil
}

func TestGetV2Impl(t *testing.T) {
	t.Parallel()
	t.Run("with client", func(t *testing.T) {
		item, err := getImpl(context.Background(), fakeClient{}, "eu-west-2b", "*")
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if item == nil {
			t.Fatalf("item is nil")
		}
		if item.Attributes.AttrStruct.Fields["zoneName"].GetStringValue() != "eu-west-2b" {
			t.Errorf("unexpected first item: %v", item)
		}
	})
}

func TestList(t *testing.T) {
	t.Parallel()
	t.Run("empty (scope mismatch)", func(t *testing.T) {
		src := AvailabilityZoneSource{}

		items, err := src.List(context.Background(), "foo.bar")
		if items != nil {
			t.Fatalf("unexpected items: %v", items)
		}
		if err == nil {
			t.Fatalf("expected err, got nil")
		}
		if !strings.HasPrefix(err.Error(), "requested scope foo.bar does not match source scope .") {
			t.Errorf("expected 'requested scope foo.bar does not match source scope .', got '%v'", err.Error())
		}
	})
}

func TestListV2Impl(t *testing.T) {
	t.Parallel()
	t.Run("with client", func(t *testing.T) {
		items, err := listImpl(context.Background(), fakeClient{}, "foo.bar")
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if len(items) != 2 {
			t.Fatalf("unexpected items (len=%v): %v", len(items), items)
		}
	})
}
