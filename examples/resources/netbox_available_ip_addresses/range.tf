resource "netbox_ip_range" "example" {
  start_address = "10.0.0.1/24"
  end_address = "10.0.0.50/24"
  description = "Example range for available IP addresses"
}

# Get multiple available IPs from an IP range
resource "netbox_available_ip_addresses" "example_range" {
  ip_range_id = netbox_ip_range.example.id
  address_count = 3
  status = "reserved"
  dns_name = "db.example.com"
  description = "Reserved from terraform"
}

# Output the allocated IP addresses
output "allocated_range_ips" {
  value = netbox_available_ip_addresses.example_range.ip_addresses
}