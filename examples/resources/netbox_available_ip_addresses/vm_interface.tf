resource "netbox_cluster_type" "example" {
  name = "Example Cluster Type"
}

resource "netbox_cluster" "example" {
  name = "example-cluster"
  cluster_type_id = netbox_cluster_type.example.id
}

resource "netbox_virtual_machine" "example" {
  name = "example-vm"
  cluster_id = netbox_cluster.example.id
  status = "active"
}

resource "netbox_interface" "example" {
  name = "eth0"
  virtual_machine_id = netbox_virtual_machine.example.id
}

resource "netbox_prefix" "example" {
  prefix = "192.168.200.0/24"
  status = "active"
}

# Allocate multiple IPs and assign them to a VM interface
resource "netbox_available_ip_addresses" "example_vm" {
  prefix_id = netbox_prefix.example.id
  address_count = 2
  status = "active"
  virtual_machine_interface_id = netbox_interface.example.id
  dns_name = "vm.example.com"
}

# Output the allocated IP addresses
output "allocated_vm_ips" {
  value = netbox_available_ip_addresses.example_vm.ip_addresses
}