package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceArcherC7DHCPReservations() *schema.Resource {
	return &schema.Resource{
		Create: resourceArcherC7DHCPReservationsCreate,
		Read:   resourceArcherC7DHCPReservationsRead,
		Update: resourceArcherC7DHCPReservationsUpdate,
		Delete: resourceArcherC7DHCPReservationsDelete,
		Importer: &schema.ResourceImporter{
			State: resourceArcherC7DHCPReservationsImportState,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},

		SchemaVersion: 1,
		MigrateState:  resourceArcherC7DHCPReservationsMigrateState,

		Schema: map[string]*schema.Schema{
			"reservations": {
				Type:       schema.TypeSet,
				Optional:   true,
				Computed:   true,
				ConfigMode: schema.SchemaConfigModeAttr,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Required: true,
						},

						"mac": {
							Type:     schema.TypeString,
							Required: true,
						},

						"ip": {
							Type:     schema.TypeString,
							Required: true,
						},

						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},

						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
				Set: resourceArcherC7DHCPReservationsRuleHash,
			},
		},
	}
}

func resourceAwsSecurityGroupCreate(d *schema.ResourceData, meta interface{}) error {
	reservations := meta.(*api.Client).DHCPAddressReservations()

	if v, ok := d.GetOk("vpc_id"); ok {
		securityGroupOpts.VpcId = aws.String(v.(string))
	}

	if v := d.Get("description"); v != nil {
		securityGroupOpts.Description = aws.String(v.(string))
	}

	groupName := naming.Generate(d.Get("name").(string), d.Get("name_prefix").(string))
	securityGroupOpts.GroupName = aws.String(groupName)

	var err error
	log.Printf(
		"[DEBUG] Security Group create configuration: %#v", securityGroupOpts)
	createResp, err := conn.CreateSecurityGroup(securityGroupOpts)
	if err != nil {
		return fmt.Errorf("Error creating Security Group: %s", err)
	}

	d.SetId(*createResp.GroupId)

	log.Printf("[INFO] Security Group ID: %s", d.Id())

	// Wait for the security group to truly exist
	resp, err := waitForSgToExist(conn, d.Id(), d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return fmt.Errorf(
			"Error waiting for Security Group (%s) to become available: %s",
			d.Id(), err)
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		if err := keyvaluetags.Ec2CreateTags(conn, d.Id(), v); err != nil {
			return fmt.Errorf("error adding EC2 Security Group (%s) tags: %s", d.Id(), err)
		}
	}

	// AWS defaults all Security Groups to have an ALLOW ALL egress rule. Here we
	// revoke that rule, so users don't unknowingly have/use it.
	group := resp.(*ec2.SecurityGroup)
	if group.VpcId != nil && *group.VpcId != "" {
		log.Printf("[DEBUG] Revoking default egress rule for Security Group for %s", d.Id())

		req := &ec2.RevokeSecurityGroupEgressInput{
			GroupId: createResp.GroupId,
			IpPermissions: []*ec2.IpPermission{
				{
					FromPort: aws.Int64(int64(0)),
					ToPort:   aws.Int64(int64(0)),
					IpRanges: []*ec2.IpRange{
						{
							CidrIp: aws.String("0.0.0.0/0"),
						},
					},
					IpProtocol: aws.String("-1"),
				},
			},
		}

		if _, err = conn.RevokeSecurityGroupEgress(req); err != nil {
			return fmt.Errorf(
				"Error revoking default egress rule for Security Group (%s): %s",
				d.Id(), err)
		}

		log.Printf("[DEBUG] Revoking default IPv6 egress rule for Security Group for %s", d.Id())
		req = &ec2.RevokeSecurityGroupEgressInput{
			GroupId: createResp.GroupId,
			IpPermissions: []*ec2.IpPermission{
				{
					FromPort: aws.Int64(int64(0)),
					ToPort:   aws.Int64(int64(0)),
					Ipv6Ranges: []*ec2.Ipv6Range{
						{
							CidrIpv6: aws.String("::/0"),
						},
					},
					IpProtocol: aws.String("-1"),
				},
			},
		}

		_, err = conn.RevokeSecurityGroupEgress(req)
		if err != nil {
			//If we have a NotFound or InvalidParameterValue, then we are trying to remove the default IPv6 egress of a non-IPv6
			//enabled SG
			if ec2err, ok := err.(awserr.Error); ok && ec2err.Code() != "InvalidPermission.NotFound" && !isAWSErr(err, "InvalidParameterValue", "remote-ipv6-range") {
				return fmt.Errorf(
					"Error revoking default IPv6 egress rule for Security Group (%s): %s",
					d.Id(), err)
			}
		}

	}

	return resourceAwsSecurityGroupUpdate(d, meta)
}
