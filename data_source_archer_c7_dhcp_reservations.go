package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	api "github.com/waisbrot/tp-link-api/lib"
)

func dataSourcecDHCPReservations() *schema.Resource {
	return &schema.Resource{
		Read: dataSourcecDHCPReservationsRead,
		Schema: map[string]*schema.Schema{

		}
	}
}
