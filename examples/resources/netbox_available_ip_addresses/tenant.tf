resource "netbox_tenant" "example" {
  name = "Example Tenant"
}

resource "netbox_ip_range" "example" {
  start_address = "192.168.50.1/24"
  end_address = "192.168.50.50/24"
  tenant_id = netbox_tenant.example.id
  description = "Example range for tenant"
}

# Get multiple available IPs from an IP range associated with a tenant
resource "netbox_available_ip_addresses" "example_tenant" {
  ip_range_id = netbox_ip_range.example.id
  address_count = 3
  status = "active"
  tenant_id = netbox_tenant.example.id
  dns_name = "tenant-service.example.com"
}

# Output the allocated IP addresses
output "allocated_tenant_ips" {
  value = netbox_available_ip_addresses.example_tenant.ip_addresses
}