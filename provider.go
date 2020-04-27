package main

import (
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func Provider() terraform.ResourceProvider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"host": {
				Type:     schema.TypeString,
				Required: true,
			},
			"username": {
				Type:     schema.TypeString,
				Required: true,
			},
			"password": {
				Type:     schema.TypeString,
				Required: true,
			},
			"router": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		DataSourcesMap: map[string]*schema.Resource{},
		ResourcesMap: map[string]*schema.Resource{
			"archer_c7_dhcp_reservations": resourceArcherC7DHCPReservations(),
		},
	}
	p.ConfigureFunc = func(data *schema.ResourceData) (interface{}, error) {
		username := data.Get("username").(string)
		password := data.Get("password").(string)
		host := data.Get("host").(string)

		log.Println("[INFO] Initializing PagerDuty client")
		return Client(username, password, host)
	}
	return p
}
