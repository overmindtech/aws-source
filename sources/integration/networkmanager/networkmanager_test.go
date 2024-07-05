package networkmanager

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/aws-source/sources/integration"
	"github.com/overmindtech/aws-source/sources/networkmanager"
	"github.com/overmindtech/sdp-go"
)

func TestNetworkManager(t *testing.T) {
	ctx := context.Background()

	t.Logf("Running NetworkManager integration tests")

	networkManagerCli, err := createNetworkManagerClient(ctx)
	if err != nil {
		t.Fatalf("failed to create NetworkManager client: %v", err)
	}

	awsCfg, err := integration.AWSSettings(ctx)
	if err != nil {
		t.Fatalf("failed to get AWS settings: %v", err)
	}

	globalNetworkSource := networkmanager.NewGlobalNetworkSource(networkManagerCli, awsCfg.AccountID)
	if globalNetworkSource.Validate() != nil {
		t.Fatalf("failed to validate NetworkManager global network source: %v", err)
	}

	siteSource := networkmanager.NewSiteSource(networkManagerCli, awsCfg.AccountID)
	if siteSource.Validate() != nil {
		t.Fatalf("failed to validate NetworkManager site source: %v", err)
	}

	linkSource := networkmanager.NewLinkSource(networkManagerCli, awsCfg.AccountID)
	if linkSource.Validate() != nil {
		t.Fatalf("failed to validate NetworkManager link source: %v", err)
	}

	linkAssociationSource := networkmanager.NewLinkAssociationSource(networkManagerCli, awsCfg.AccountID)
	if linkAssociationSource.Validate() != nil {
		t.Fatalf("failed to validate NetworkManager link association source: %v", err)
	}

	connectionSource := networkmanager.NewConnectionSource(networkManagerCli, awsCfg.AccountID)
	if connectionSource.Validate() != nil {
		t.Fatalf("failed to validate NetworkManager connection source: %v", err)
	}

	deviceSource := networkmanager.NewDeviceSource(networkManagerCli, awsCfg.AccountID)
	if deviceSource.Validate() != nil {
		t.Fatalf("failed to validate NetworkManager device source: %v", err)
	}

	globalScope := sources.FormatScope(awsCfg.AccountID, "")

	t.Run("Global Network", func(t *testing.T) {
		// List global networks
		globalNetworks, err := globalNetworkSource.List(ctx, globalScope, true)
		if err != nil {
			t.Fatalf("failed to list NetworkManager global networks: %v", err)
		}

		if len(globalNetworks) == 0 {
			t.Fatalf("no global networks found")
		}

		globalNetworkUniqueAttribute := globalNetworks[0].GetUniqueAttribute()

		globalNetworkID, err := integration.GetUniqueAttributeValue(
			globalNetworkUniqueAttribute,
			globalNetworks,
			integration.ResourceTags(integration.NetworkManager, globalNetworkSrc),
		)
		if err != nil {
			t.Fatalf("failed to get global network ID: %v", err)
		}

		// Get global network
		globalNetwork, err := globalNetworkSource.Get(ctx, globalScope, globalNetworkID, true)
		if err != nil {
			t.Fatalf("failed to get NetworkManager global network: %v", err)
		}

		globalNetworkIDFromGet, err := integration.GetUniqueAttributeValue(
			globalNetworkUniqueAttribute,
			[]*sdp.Item{globalNetwork},
			integration.ResourceTags(integration.NetworkManager, globalNetworkSrc),
		)
		if err != nil {
			t.Fatalf("failed to get global network ID from get: %v", err)
		}

		if globalNetworkID != globalNetworkIDFromGet {
			t.Fatalf("expected global network ID %s, got %s", globalNetworkID, globalNetworkIDFromGet)
		}

		// Search global network by ARN
		globalNetworkARN, err := globalNetwork.GetAttributes().Get("globalNetworkArn")
		if err != nil {
			t.Fatalf("failed to get global network ARN: %v", err)
		}

		if globalScope != globalNetwork.GetScope() {
			t.Fatalf("expected global scope %s, got %s", globalScope, globalNetwork.GetScope())
		}

		globalNetworks, err = globalNetworkSource.Search(ctx, globalScope, globalNetworkARN.(string), true)
		if err != nil {
			t.Fatalf("failed to search NetworkManager global networks: %v", err)
		}

		if len(globalNetworks) == 0 {
			t.Fatalf("no global networks found")
		}

		globalNetworkIDFromSearch, err := integration.GetUniqueAttributeValue(
			globalNetworkUniqueAttribute,
			globalNetworks,
			integration.ResourceTags(integration.NetworkManager, globalNetworkSrc),
		)
		if err != nil {
			t.Fatalf("failed to get global network ID from search: %v", err)
		}

		if globalNetworkID != globalNetworkIDFromSearch {
			t.Fatalf("expected global network ID %s, got %s", globalNetworkID, globalNetworkIDFromSearch)
		}

		t.Run("Site", func(t *testing.T) {
			// Search sites by the global network ID that they are created on
			sites, err := siteSource.Search(ctx, globalScope, globalNetworkID, true)
			if err != nil {
				t.Fatalf("failed to search for site: %v", err)
			}

			if len(sites) == 0 {
				t.Fatalf("no sites found")
			}

			siteUniqueAttribute := sites[0].GetUniqueAttribute()

			// composite site id is in the format of {globalNetworkID}|{siteID}
			compositeSiteID, err := integration.GetUniqueAttributeValue(
				siteUniqueAttribute,
				sites,
				integration.ResourceTags(integration.NetworkManager, siteSrc),
			)
			if err != nil {
				t.Fatalf("failed to get site ID from search: %v", err)
			}

			// Get site: query format = globalNetworkID|siteID
			site, err := siteSource.Get(ctx, globalScope, compositeSiteID, true)
			if err != nil {
				t.Fatalf("failed to get site: %v", err)
			}

			siteIDFromGet, err := integration.GetUniqueAttributeValue(
				siteUniqueAttribute,
				[]*sdp.Item{site},
				integration.ResourceTags(integration.NetworkManager, siteSrc),
			)
			if err != nil {
				t.Fatalf("failed to get site ID from get: %v", err)
			}

			if compositeSiteID != siteIDFromGet {
				t.Fatalf("expected site ID %s, got %s", compositeSiteID, siteIDFromGet)
			}

			siteID := strings.Split(compositeSiteID, "|")[1]

			t.Run("Link", func(t *testing.T) {
				// Search links by the global network ID that they are created on
				links, err := linkSource.Search(ctx, globalScope, globalNetworkID, true)
				if err != nil {
					t.Fatalf("failed to search for link: %v", err)
				}

				if len(links) == 0 {
					t.Fatalf("no links found")
				}

				linkUniqueAttribute := links[0].GetUniqueAttribute()

				compositeLinkID, err := integration.GetUniqueAttributeValue(
					linkUniqueAttribute,
					links,
					integration.ResourceTags(integration.NetworkManager, linkSrc),
				)
				if err != nil {
					t.Fatalf("failed to get link ID from search: %v", err)
				}

				// Get link: query format = globalNetworkID|linkID
				link, err := linkSource.Get(ctx, globalScope, compositeLinkID, true)
				if err != nil {
					t.Fatalf("failed to get link: %v", err)
				}

				linkIDFromGet, err := integration.GetUniqueAttributeValue(
					linkUniqueAttribute,
					[]*sdp.Item{link},
					integration.ResourceTags(integration.NetworkManager, linkSrc),
				)

				if compositeLinkID != linkIDFromGet {
					t.Fatalf("expected link ID %s, got %s", compositeLinkID, linkIDFromGet)
				}

				linkID := strings.Split(compositeLinkID, "|")[1]

				t.Run("Device", func(t *testing.T) {
					// Search devices by the global network ID and site ID
					// query format = globalNetworkID|siteID
					queryDevice := fmt.Sprintf("%s|%s", globalNetworkID, siteID)
					devices, err := deviceSource.Search(ctx, globalScope, queryDevice, true)
					if err != nil {
						t.Fatalf("failed to search for device: %v", err)
					}

					if len(devices) == 0 {
						t.Fatalf("no devices found")
					}

					deviceUniqueAttribute := devices[0].GetUniqueAttribute()

					// composite device id is in the format of: {globalNetworkID}|{deviceID}
					deviceOneCompositeID, err := integration.GetUniqueAttributeValue(
						deviceUniqueAttribute,
						devices,
						integration.ResourceTags(integration.NetworkManager, deviceSrc, deviceOneName),
					)
					if err != nil {
						t.Fatalf("failed to get device ID from search: %v", err)
					}

					// Get device: query format = globalNetworkID|deviceID
					device, err := deviceSource.Get(ctx, globalScope, deviceOneCompositeID, true)
					if err != nil {
						t.Fatalf("failed to get device: %v", err)
					}

					deviceOneCompositeIDFromGet, err := integration.GetUniqueAttributeValue(
						deviceUniqueAttribute,
						[]*sdp.Item{device},
						integration.ResourceTags(integration.NetworkManager, deviceSrc, deviceOneName),
					)
					if err != nil {
						t.Fatalf("failed to get device ID from get: %v", err)
					}

					if deviceOneCompositeID != deviceOneCompositeIDFromGet {
						t.Fatalf("expected device ID %s, got %s", deviceOneCompositeID, deviceOneCompositeIDFromGet)
					}

					deviceOneID := strings.Split(deviceOneCompositeID, "|")[1]

					// Search devices by the global network ID
					devicesByGlobalNetwork, err := deviceSource.Search(ctx, globalScope, globalNetworkID, true)
					if err != nil {
						t.Fatalf("failed to search for device by global network: %v", err)
					}

					integration.AssertEqualItems(t, devices, devicesByGlobalNetwork, deviceUniqueAttribute)

					t.Run("Link Association", func(t *testing.T) {
						// Search link associations by the global network ID, link ID
						queryLALink := fmt.Sprintf("%s|link|%s", globalNetworkID, linkID)
						linkAssociations, err := linkAssociationSource.Search(ctx, globalScope, queryLALink, true)
						if err != nil {
							t.Fatalf("failed to search for link association: %v", err)
						}

						if len(linkAssociations) == 0 {
							t.Fatalf("no link associations found")
						}

						linkAssociationUniqueAttribute := linkAssociations[0].GetUniqueAttribute()

						// composite link association id is in the format of: {globalNetworkID}|{linkID}|{deviceID}
						compositeLinkAssociationID, err := integration.GetUniqueAttributeValue(
							linkAssociationUniqueAttribute,
							linkAssociations,
							nil, // we didn't use tags on associations
						)
						if err != nil {
							t.Fatalf("failed to get link association ID from search: %v", err)
						}

						// Get link association: query format = globalNetworkID|linkID|deviceID
						linkAssociation, err := linkAssociationSource.Get(ctx, globalScope, compositeLinkAssociationID, true)
						if err != nil {
							t.Fatalf("failed to get link association: %v", err)
						}

						compositeLinkAssociationIDFromGet, err := integration.GetUniqueAttributeValue(
							linkAssociationUniqueAttribute,
							[]*sdp.Item{linkAssociation},
							nil, // we didn't use tags on associations
						)

						if compositeLinkAssociationID != compositeLinkAssociationIDFromGet {
							t.Fatalf("expected link association ID %s, got %s", compositeLinkAssociationID, compositeLinkAssociationIDFromGet)
						}

						// Search link associations by the global network ID
						searchLinkAssociationsByGlobalNetwork, err := linkAssociationSource.Search(ctx, globalScope, globalNetworkID, true)
						if err != nil {
							t.Fatalf("failed to search for link association by global network: %v", err)
						}

						integration.AssertEqualItems(t, linkAssociations, searchLinkAssociationsByGlobalNetwork, linkAssociationUniqueAttribute)

						// Search link associations by the global network ID and device ID
						queryLADevice := fmt.Sprintf("%s|device|%s", globalNetworkID, deviceOneID)
						linkAssociationsByDevice, err := linkAssociationSource.Search(ctx, globalScope, queryLADevice, true)
						if err != nil {
							t.Fatalf("failed to search for link association by device: %v", err)
						}

						integration.AssertEqualItems(t, linkAssociations, linkAssociationsByDevice, linkAssociationUniqueAttribute)
					})

					t.Run("Connection", func(t *testing.T) {
						// Search connections by the global network ID
						connections, err := connectionSource.Search(ctx, globalScope, globalNetworkID, true)
						if err != nil {
							t.Fatalf("failed to search for connection: %v", err)
						}

						if len(connections) == 0 {
							t.Fatalf("no connections found")
						}

						connectionUniqueAttribute := connections[0].GetUniqueAttribute()

						// composite connection id is in the format of: {globalNetworkID}|{connectionID}
						compositeConnectionID, err := integration.GetUniqueAttributeValue(
							connectionUniqueAttribute,
							connections,
							nil, // we didn't use tags on connections
						)
						if err != nil {
							t.Fatalf("failed to get connection ID from search: %v", err)
						}

						// Get connection: query format = globalNetworkID|connectionID
						connection, err := connectionSource.Get(ctx, globalScope, compositeConnectionID, true)
						if err != nil {
							t.Fatalf("failed to get connection: %v", err)
						}

						compositeConnectionIDFromGet, err := integration.GetUniqueAttributeValue(
							connectionUniqueAttribute,
							[]*sdp.Item{connection},
							nil, // we didn't use tags on connections
						)

						if compositeConnectionID != compositeConnectionIDFromGet {
							t.Fatalf("expected connection ID %s, got %s", compositeConnectionID, compositeConnectionIDFromGet)
						}

						// Search connections by global network ID and device ID
						queryCon := fmt.Sprintf("%s|%s", globalNetworkID, deviceOneID)
						connectionsByDevice, err := connectionSource.Search(ctx, globalScope, queryCon, true)
						if err != nil {
							t.Fatalf("failed to search for connection by device: %v", err)
						}

						integration.AssertEqualItems(t, connections, connectionsByDevice, connectionUniqueAttribute)
					})
				})
			})
		})
	})
}
