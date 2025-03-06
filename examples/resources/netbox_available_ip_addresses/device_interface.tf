resource "netbox_site" "example" {
  name = "Example Site"
  status = "active"
}

resource "netbox_device_role" "example" {
  name = "Server"
  color = "ff0000"
}

resource "netbox_manufacturer" "example" {
  name = "Example Manufacturer"
}

resource "netbox_device_type" "example" {
  model = "Example Device Type"
  manufacturer_id = netbox_manufacturer.example.id
}

resource "netbox_device" "example" {
  name = "example-device"
  device_type_id = netbox_device_type.example.id
  role_id = netbox_device_role.example.id
  site_id = netbox_site.example.id
  status = "active"
}

resource "netbox_device_interface" "example" {
  name = "eth0"
  device_id = netbox_device.example.id
  type = "1000base-t"
}

resource "netbox_ip_range" "example" {
  start_address = "172.16.0.1/24"
  end_address = "172.16.0.50/24"
  description = "Example range for device interfaces"
}

# Allocate multiple IPs and assign them to a device interface
resource "netbox_available_ip_addresses" "example_interface" {
  ip_range_id = netbox_ip_range.example.id
  address_count = 2
  status = "active"
  device_interface_id = netbox_device_interface.example.id
  dns_name = "server-interfaces.example.com"
}

# Output the allocated IP addresses
output "allocated_interface_ips" {
  value = netbox_available_ip_addresses.example_interface.ip_addresses
}