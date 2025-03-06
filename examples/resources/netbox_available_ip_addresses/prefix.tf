resource "netbox_prefix" "example" {
  prefix = "192.168.100.0/24"
  status = "active"
}

# Get multiple available IPs from a prefix
resource "netbox_available_ip_addresses" "example" {
  prefix_id = netbox_prefix.example.id
  address_count = 5
  status = "active"
  dns_name = "server.example.com"
  description = "Allocated from terraform"
  role = "loopback"
}

# Output the allocated IP addresses
output "allocated_ips" {
  value = netbox_available_ip_addresses.example.ip_addresses
}