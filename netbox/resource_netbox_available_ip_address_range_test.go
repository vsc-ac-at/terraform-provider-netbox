package netbox

import (
	"fmt"
	"log"
	"testing"

	"github.com/fbreckle/go-netbox/netbox/client"
	"github.com/fbreckle/go-netbox/netbox/client/ipam"
	"github.com/fbreckle/go-netbox/netbox/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNetboxAvailableIPAddressRange_basic_prefix(t *testing.T) {
	testPrefix := "2.2.2.0/24"
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_prefix" "test" {
  prefix = "%s"
  status = "active"
  is_pool = false
}
resource "netbox_available_ip_addresses" "test" {
  prefix_id = netbox_prefix.test.id
  address_count = 3
  status = "active"
  dns_name = "test.mydomain.local"
  role = "loopback"
}`, testPrefix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("netbox_available_ip_addresses.test", "ip_addresses.0"),
					resource.TestCheckResourceAttrSet("netbox_available_ip_addresses.test", "ip_addresses.1"),
					resource.TestCheckResourceAttrSet("netbox_available_ip_addresses.test", "ip_addresses.2"),
					resource.TestCheckResourceAttr("netbox_available_ip_addresses.test", "address_count", "3"),
					resource.TestCheckResourceAttr("netbox_available_ip_addresses.test", "status", "active"),
					resource.TestCheckResourceAttr("netbox_available_ip_addresses.test", "dns_name", "test.mydomain.local"),
					resource.TestCheckResourceAttr("netbox_available_ip_addresses.test", "role", "loopback"),
				),
			},
		},
	})
}

func TestAccNetboxAvailableIPAddressRange_basic_range(t *testing.T) {
	startAddress := "2.2.5.1/24"
	endAddress := "2.2.5.50/24"
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_ip_range" "test" {
  start_address = "%s"
  end_address = "%s"
}
resource "netbox_available_ip_addresses" "test_range" {
  ip_range_id = netbox_ip_range.test.id
  address_count = 3
  status = "active"
  dns_name = "test_range.mydomain.local"
}`, startAddress, endAddress),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("netbox_available_ip_addresses.test_range", "ip_addresses.0"),
					resource.TestCheckResourceAttrSet("netbox_available_ip_addresses.test_range", "ip_addresses.1"),
					resource.TestCheckResourceAttrSet("netbox_available_ip_addresses.test_range", "ip_addresses.2"),
					resource.TestCheckResourceAttr("netbox_available_ip_addresses.test_range", "address_count", "3"),
					resource.TestCheckResourceAttr("netbox_available_ip_addresses.test_range", "status", "active"),
					resource.TestCheckResourceAttr("netbox_available_ip_addresses.test_range", "dns_name", "test_range.mydomain.local"),
				),
			},
		},
	})
}

func TestAccNetboxAvailableIPAddressRange_with_vrf(t *testing.T) {
	testPrefix := "3.3.3.0/24"
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_prefix" "test" {
  prefix = "%s"
  status = "active"
  is_pool = false
}
resource "netbox_vrf" "test" {
  name = "test_vrf_for_ip_range"
}
resource "netbox_available_ip_addresses" "test" {
  prefix_id = netbox_prefix.test.id
  address_count = 2
  status = "active"
  vrf_id = netbox_vrf.test.id
}`, testPrefix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("netbox_available_ip_addresses.test", "ip_addresses.0"),
					resource.TestCheckResourceAttrSet("netbox_available_ip_addresses.test", "ip_addresses.1"),
					resource.TestCheckResourceAttr("netbox_available_ip_addresses.test", "address_count", "2"),
					resource.TestCheckResourceAttrPair("netbox_available_ip_addresses.test", "vrf_id", "netbox_vrf.test", "id"),
				),
			},
		},
	})
}

func TestAccNetboxAvailableIPAddressRange_with_tenant(t *testing.T) {
	testPrefix := "3.4.3.0/24"
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_prefix" "test" {
  prefix = "%s"
  status = "active"
  is_pool = false
}
resource "netbox_tenant" "test" {
  name = "test_tenant_for_ip_range"
}
resource "netbox_available_ip_addresses" "test" {
  prefix_id = netbox_prefix.test.id
  address_count = 2
  status = "active"
  tenant_id = netbox_tenant.test.id
}`, testPrefix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("netbox_available_ip_addresses.test", "ip_addresses.0"),
					resource.TestCheckResourceAttrSet("netbox_available_ip_addresses.test", "ip_addresses.1"),
					resource.TestCheckResourceAttr("netbox_available_ip_addresses.test", "address_count", "2"),
					resource.TestCheckResourceAttrPair("netbox_available_ip_addresses.test", "tenant_id", "netbox_tenant.test", "id"),
				),
			},
		},
	})
}

func TestAccNetboxAvailableIPAddressRange_device_interface(t *testing.T) {
	startAddress := "3.5.5.1/24"
	endAddress := "3.5.5.50/24"
	testSlug := "av_ipa_range_dev"
	testName := testAccGetTestName(testSlug)
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNetboxIPAddressFullDeviceDependencies(testName) + fmt.Sprintf(`
resource "netbox_ip_range" "test_range" {
  start_address = "%s"
  end_address = "%s"
}
resource "netbox_available_ip_addresses" "test" {
  ip_range_id = netbox_ip_range.test_range.id
  address_count = 2
  status = "active"
  dns_name = "test_range.mydomain.local"
  device_interface_id = netbox_device_interface.test.id
}`, startAddress, endAddress),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("netbox_available_ip_addresses.test", "ip_addresses.0"),
					resource.TestCheckResourceAttrSet("netbox_available_ip_addresses.test", "ip_addresses.1"),
					resource.TestCheckResourceAttr("netbox_available_ip_addresses.test", "status", "active"),
					resource.TestCheckResourceAttrPair("netbox_available_ip_addresses.test", "device_interface_id", "netbox_device_interface.test", "id"),
				),
			},
		},
	})
}

func TestAccNetboxAvailableIPAddressRange_vm_interface(t *testing.T) {
	startAddress := "3.6.5.1/24"
	endAddress := "3.6.5.50/24"
	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "netbox_ip_range" "test_range" {
  start_address = "%s"
  end_address = "%s"
}
resource "netbox_cluster_type" "test" {
  name = "test_cluster_type_for_ip_range"
}
resource "netbox_cluster" "test" {
  name = "test_cluster_for_ip_range"
  cluster_type_id = netbox_cluster_type.test.id
}
resource "netbox_virtual_machine" "test" {
  name = "test_vm_for_ip_range"
  cluster_id = netbox_cluster.test.id
}
resource "netbox_interface" "test" {
  name = "test_vm_interface_for_ip_range"
  virtual_machine_id = netbox_virtual_machine.test.id
}
resource "netbox_available_ip_addresses" "test" {
  ip_range_id = netbox_ip_range.test_range.id
  address_count = 2
  status = "active"
  dns_name = "test_range.mydomain.local"
  virtual_machine_interface_id = netbox_interface.test.id
}`, startAddress, endAddress),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("netbox_available_ip_addresses.test", "ip_addresses.0"),
					resource.TestCheckResourceAttrSet("netbox_available_ip_addresses.test", "ip_addresses.1"),
					resource.TestCheckResourceAttr("netbox_available_ip_addresses.test", "status", "active"),
					resource.TestCheckResourceAttrPair("netbox_available_ip_addresses.test", "virtual_machine_interface_id", "netbox_interface.test", "id"),
				),
			},
		},
	})
}

func init() {
	resource.AddTestSweepers("netbox_available_ip_addresses", &resource.Sweeper{
		Name:         "netbox_available_ip_addresses",
		Dependencies: []string{},
		F: func(region string) error {
			m, err := sharedClientForRegion(region)
			if err != nil {
				return fmt.Errorf("Error getting client: %s", err)
			}
			api := m.(*client.NetBoxAPI)
			params := ipam.NewIpamIPAddressesListParams()
			res, err := api.Ipam.IpamIPAddressesList(params, nil)
			if err != nil {
				return err
			}
			for _, ipAddress := range res.GetPayload().Results {
				if len(ipAddress.Tags) > 0 && (ipAddress.Tags[0] == &models.NestedTag{Name: strToPtr("acctest"), Slug: strToPtr("acctest")}) {
					deleteParams := ipam.NewIpamIPAddressesDeleteParams().WithID(ipAddress.ID)
					_, err := api.Ipam.IpamIPAddressesDelete(deleteParams, nil)
					if err != nil {
						return err
					}
					log.Print("[DEBUG] Deleted an ip address")
				}
			}
			return nil
		},
	})
}