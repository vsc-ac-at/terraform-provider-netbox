package netbox

import (
	"errors"
	"fmt"

	"github.com/fbreckle/go-netbox/netbox/client"
	"github.com/fbreckle/go-netbox/netbox/client/ipam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxServices() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxServicesListRead,
		Description: `:meta:subcategory:Extras:`,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"filter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"limit": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"services": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"virtual_machine_id": {
							Type:     schema.TypeInt,
							Computed: true,				
						},			
						"protocol": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ports": {
							Type:         schema.TypeSet,
							Computed:     true,
							Elem: &schema.Schema{
								Type: schema.TypeInt,
							},
						},
						"ip_addresses": {
							Type:         schema.TypeList,
							Computed:     true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"address": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"tags": {
							Type: schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"tag_id": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"slug": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						customFieldsKey: customFieldsSchema,
					},
				},
			},
		},
	}
}

func dataSourceNetboxServicesListRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*client.NetBoxAPI)

	params := ipam.NewIpamServicesListParams()

	name := d.Get("name").(string)
	params.Name = &name

	if limit, ok := d.GetOk("limit"); ok {
		limitInt := int64(limit.(int))
		params.Limit = &limitInt
	}
	
	if filter, ok := d.GetOk("filter"); ok {
		var filterParams = filter.(*schema.Set)
		var tags []string
		for _, f := range filterParams.List() {
			k := f.(map[string]interface{})["name"]
			v := f.(map[string]interface{})["value"]
			vString := v.(string)
			switch k {
			case "protocol":
				params.Protocol = &vString
			case "tag":
				tags = append(tags, vString)
				params.Tag = tags
			default:
				return fmt.Errorf("'%s' is not a supported filter parameter", k)
			}
		}
	}

	res, err := api.Ipam.IpamServicesList(params, nil)
	if err != nil {
		if errresp, ok := err.(*ipam.IpamServicesListDefault); ok {
			errorcode := errresp.Code()
			if errorcode == 404 {
				// If the ID is updated to blank, this tells Terraform the resource no longer exists (maybe it was destroyed out of band). Just like the destroy callback, the Read function should gracefully handle this case. https://www.terraform.io/docs/extend/writing-custom-providers.html
				d.SetId("")
				return nil
			}
		}
		return err
	}

	if *res.GetPayload().Count == int64(0) {
		return errors.New("no service found matching filter")
	}

	var services []map[string]interface{}
	for _, v := range res.GetPayload().Results {
		var s = make(map[string]interface{})

		s["id"] = v.ID
		s["name"] = v.Name
		if v.VirtualMachine != nil {
			s["virtual_machine_id"] = v.VirtualMachine.ID	
		}	
		s["protocol"] = v.Protocol.Value
		s["ports"] = v.Ports
		s["description"] = v.Description
	
		var tags []map[string]interface{}
		for _, t := range v.Tags {
			mapping := make(map[string]interface{})
	
			mapping["tag_id"] = t.ID
			mapping["name"] = t.Name
			mapping["slug"] = t.Slug
	
			tags = append(tags, mapping)
		}
		s["tags"] = tags
	
		var ip_addresses []map[string]interface{}
		for _, ip := range v.Ipaddresses {
			mapping := make(map[string]interface{})
	
			mapping["id"] = ip.ID
			mapping["address"] = ip.Address
	
			ip_addresses = append(ip_addresses, mapping)
		}
		s["ip_addresses"] = ip_addresses
	
		cf := getCustomFields(v.CustomFields)
		if cf != nil {
			s[customFieldsKey] = cf
		}
		
		services = append(services, s)
	}

	d.SetId(id.UniqueId())
	return d.Set("services", services)
}
