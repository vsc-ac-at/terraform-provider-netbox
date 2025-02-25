package netbox

import (
	"strconv"

	"github.com/fbreckle/go-netbox/netbox/client"
	"github.com/fbreckle/go-netbox/netbox/client/ipam"
	"github.com/fbreckle/go-netbox/netbox/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxAvailableIPAddressRange() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxAvailableIPAddressRangeCreate,
		Read:   resourceNetboxAvailableIPAddressRangeRead,
		Update: resourceNetboxAvailableIPAddressRangeUpdate,
		Delete: resourceNetboxAvailableIPAddressRangeDelete,

		Description: `:meta:subcategory:IP Address Management (IPAM):Per [the docs](https://netbox.readthedocs.io/en/stable/models/ipam/ipaddress/):

> An IP address comprises a single host address (either IPv4 or IPv6) and its subnet mask. Its mask should match exactly how the IP address is configured on an interface in the real world.
> Like a prefix, an IP address can optionally be assigned to a VRF (otherwise, it will appear in the "global" table). IP addresses are automatically arranged under parent prefixes within their respective VRFs according to the IP hierarchya.
>
> Each IP address can also be assigned an operational status and a functional role. Statuses are hard-coded in NetBox and include the following:
> * Active
> * Reserved
> * Deprecated
> * DHCP
> * SLAAC (IPv6 Stateless Address Autoconfiguration)

This resource will retrieve the next available IP addresses from a given prefix or IP range (specified by ID)`,

		Schema: map[string]*schema.Schema{
			"prefix_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				ExactlyOneOf: []string{"prefix_id", "ip_range_id"},
			},
			"ip_range_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				ExactlyOneOf: []string{"prefix_id", "ip_range_id"},
			},
			"ip_addresses": {
				Type:     schema.TypeList,
				Computed: true,
			},
			"interface_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				RequiredWith: []string{"object_type"},
			},
			"object_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(resourceNetboxIPAddressObjectTypeOptions, false),
				Description:  buildValidValueDescription(resourceNetboxIPAddressObjectTypeOptions),
				RequiredWith: []string{"interface_id"},
			},
			"virtual_machine_interface_id": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"interface_id", "device_interface_id"},
			},
			"device_interface_id": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"interface_id", "virtual_machine_interface_id"},
			},
			"vrf_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"tenant_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(resourceNetboxIPAddressStatusOptions, false),
				Description:  buildValidValueDescription(resourceNetboxIPAddressStatusOptions),
				Default:      "active",
			},
			"dns_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			tagsKey: tagsSchema,
			"role": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(resourceNetboxIPAddressRoleOptions, false),
				Description:  buildValidValueDescription(resourceNetboxIPAddressRoleOptions),
			},
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceNetboxAvailableIPAddressRangeCreate(d *schema.ResourceData, m interface{}) error {
	api := m.(*client.NetBoxAPI)
	prefixID := int64(d.Get("prefix_id").(int))
	vrfID := int64(int64(d.Get("vrf_id").(int)))
	rangeID := int64(d.Get("ip_range_id").(int))
	nestedvrf := models.NestedVRF{
		ID: vrfID,
	}
	data := models.AvailableIP{
		Vrf: &nestedvrf,
	}
	if prefixID != 0 {
		params := ipam.NewIpamPrefixesAvailableIpsCreateParams().WithID(prefixID).WithData([]*models.AvailableIP{&data})
		res, _ := api.Ipam.IpamPrefixesAvailableIpsCreate(params, nil)
		// Since we generated the ip_address, set that now
		d.SetId(strconv.FormatInt(res.Payload[0].ID, 10))
		d.Set("ip_address", *res.Payload[0].Address)
	}
	if rangeID != 0 {
		params := ipam.NewIpamIPRangesAvailableIpsCreateParams().WithID(rangeID).WithData([]*models.AvailableIP{&data})
		res, _ := api.Ipam.IpamIPRangesAvailableIpsCreate(params, nil)
		// Since we generated the ip_address, set that now
		d.SetId(strconv.FormatInt(res.Payload[0].ID, 10))
		d.Set("ip_address", *res.Payload[0].Address)
	}
	return resourceNetboxAvailableIPAddressRangeUpdate(d, m)
}

func resourceNetboxAvailableIPAddressRangeRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*client.NetBoxAPI)
	id, _ := strconv.ParseInt(d.Id(), 10, 64)
	params := ipam.NewIpamIPAddressesReadParams().WithID(id)

	res, err := api.Ipam.IpamIPAddressesRead(params, nil)
	if err != nil {
		if errresp, ok := err.(*ipam.IpamIPAddressesReadDefault); ok {
			errorcode := errresp.Code()
			if errorcode == 404 {
				// If the ID is updated to blank, this tells Terraform the resource no longer exists (maybe it was destroyed out of band). Just like the destroy callback, the Read function should gracefully handle this case. https://www.terraform.io/docs/extend/writing-custom-providers.html
				d.SetId("")
				return nil
			}
		}
		return err
	}

	ipAddress := res.GetPayload()
	if ipAddress.AssignedObjectID != nil {
		vmInterfaceID := getOptionalInt(d, "virtual_machine_interface_id")
		deviceInterfaceID := getOptionalInt(d, "device_interface_id")
		interfaceID := getOptionalInt(d, "interface_id")

		switch {
		case vmInterfaceID != nil:
			d.Set("virtual_machine_interface_id", ipAddress.AssignedObjectID)
		case deviceInterfaceID != nil:
			d.Set("device_interface_id", ipAddress.AssignedObjectID)
		// if interfaceID is given, object_type must be set as well
		case interfaceID != nil:
			d.Set("object_type", ipAddress.AssignedObjectType)
			d.Set("interface_id", ipAddress.AssignedObjectID)
		}
	} else {
		d.Set("interface_id", nil)
		d.Set("object_type", "")
	}

	if ipAddress.Vrf != nil {
		d.Set("vrf_id", ipAddress.Vrf.ID)
	} else {
		d.Set("vrf_id", nil)
	}

	if ipAddress.Tenant != nil {
		d.Set("tenant_id", ipAddress.Tenant.ID)
	} else {
		d.Set("tenant_id", nil)
	}

	if ipAddress.DNSName != "" {
		d.Set("dns_name", ipAddress.DNSName)
	}

	d.Set("ip_address", ipAddress.Address)
	d.Set("description", ipAddress.Description)
	d.Set("status", ipAddress.Status.Value)
	d.Set(tagsKey, getTagListFromNestedTagList(ipAddress.Tags))

	cf := getCustomFields(ipAddress.CustomFields)
	if cf != nil {
		d.Set(customFieldsKey, cf)
	}

	return nil
}

func resourceNetboxAvailableIPAddressRangeUpdate(d *schema.ResourceData, m interface{}) error {
	api := m.(*client.NetBoxAPI)

	id, _ := strconv.ParseInt(d.Id(), 10, 64)
	data := models.WritableIPAddress{}

	data.Address = strToPtr(d.Get("ip_address").(string))
	data.Status = d.Get("status").(string)

	data.Description = getOptionalStr(d, "description", false)
	data.Role = getOptionalStr(d, "role", false)
	data.DNSName = getOptionalStr(d, "dns_name", false)
	data.Vrf = getOptionalInt(d, "vrf_id")
	data.Tenant = getOptionalInt(d, "tenant_id")

	if interfaceID, ok := d.GetOk("interface_id"); ok {
		// The other possible type is dcim.interface for devices
		data.AssignedObjectType = strToPtr("virtualization.vminterface")
		data.AssignedObjectID = int64ToPtr(int64(interfaceID.(int)))
	}

	vmInterfaceID := getOptionalInt(d, "virtual_machine_interface_id")
	deviceInterfaceID := getOptionalInt(d, "device_interface_id")
	interfaceID := getOptionalInt(d, "interface_id")

	switch {
	case vmInterfaceID != nil:
		data.AssignedObjectType = strToPtr("virtualization.vminterface")
		data.AssignedObjectID = vmInterfaceID
	case deviceInterfaceID != nil:
		data.AssignedObjectType = strToPtr("dcim.interface")
		data.AssignedObjectID = deviceInterfaceID
	// if interfaceID is given, object_type must be set as well
	case interfaceID != nil:
		data.AssignedObjectType = strToPtr(d.Get("object_type").(string))
		data.AssignedObjectID = interfaceID
	// default = ip is not linked to anything
	default:
		data.AssignedObjectType = strToPtr("")
		data.AssignedObjectID = nil
	}

	data.Tags, _ = getNestedTagListFromResourceDataSet(api, d.Get(tagsKey))

	ct, ok := d.GetOk(customFieldsKey)
	if ok {
		data.CustomFields = ct
	}

	params := ipam.NewIpamIPAddressesUpdateParams().WithID(id).WithData(&data)

	_, err := api.Ipam.IpamIPAddressesUpdate(params, nil)
	if err != nil {
		return err
	}
	return resourceNetboxAvailableIPAddressRangeRead(d, m)
}

func resourceNetboxAvailableIPAddressRangeDelete(d *schema.ResourceData, m interface{}) error {
	api := m.(*client.NetBoxAPI)

	id, _ := strconv.ParseInt(d.Id(), 10, 64)
	params := ipam.NewIpamIPAddressesDeleteParams().WithID(id)

	_, err := api.Ipam.IpamIPAddressesDelete(params, nil)
	if err != nil {
		if errresp, ok := err.(*ipam.IpamIPAddressesDeleteDefault); ok {
			if errresp.Code() == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}
	return nil
}
