resource "netbox_vrf" "example" {
  name = "Example VRF"
  rd = "65000:1"
}

resource "netbox_prefix" "example" {
  prefix = "10.10.0.0/16"
  status = "active"
  vrf_id = netbox_vrf.example.id
}

# Get multiple available IPs from a prefix in a VRF
resource "netbox_available_ip_addresses" "example_vrf" {
  prefix_id = netbox_prefix.example.id
  address_count = 4
  status = "active"
  vrf_id = netbox_vrf.example.id
  dns_name = "vrf-service.example.com"
}

# Output the allocated IP addresses
output "allocated_vrf_ips" {
  value = netbox_available_ip_addresses.example_vrf.ip_addresses
}