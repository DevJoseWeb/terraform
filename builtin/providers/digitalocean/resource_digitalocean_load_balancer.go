package digitalocean

import (
	"fmt"
	"log"

	"github.com/digitalocean/godo"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceDigitalOceanLoadBalancer() *schema.Resource {
	return &schema.Resource{
		Create: resourceDigitalOceanLoadBalancerCreate,
		Read:   resourceDigitalOceanLoadBalancerRead,
		Delete: resourceDigitalOceanLoadBalancerDelete,
		Importer: &schema.ResourceImporter{
			State: resourceDigitalOceanLoadBalancerImport,
		},

		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"ip": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"algorithm": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"region": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"droplet_ids": &schema.Schema{
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeInt},
				Computed: true,
			},
		},
	}
}

func resourceDigitalOceanLoadBalancerCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*godo.Client)

	opts := &godo.LoadBalancerRequest{
		Name:      d.Get("name").(string),
		Algorithm: d.Get("algorithm").(string),
		Region:    d.Get("region").(string),
	}

	log.Printf("[DEBUG] Load Balancer create configuration: %#v", opts)
	lb, _, err := client.LoadBalancers.Create(opts)
	if err != nil {
		return fmt.Errorf("Error creating Load Balancer: %s", err)
	}

	d.SetId(lb.ID)
	log.Printf("[INFO] Load Balancer name: %s", lb.Name)

	return resourceDigitalOceanLoadBalancerRead(d, meta)
}

func resourceDigitalOceanLoadBalancerRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*godo.Client)

	lb, resp, err := client.LoadBalancers.Get(d.Id())
	if err != nil {
		// If the volume is somehow already destroyed, mark as
		// successfully gone
		if resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error retrieving Load Balancer: %s", err)
	}

	d.Set("id", lb.ID)

	dids := make([]interface{}, 0, len(lb.DropletIDs))
	for _, did := range lb.DropletIDs {
		dids = append(dids, did)
	}
	d.Set("droplet_ids", schema.NewSet(
		func(dropletID interface{}) int { return dropletID.(int) },
		dids,
	))

	return nil
}

func resourceDigitalOceanLoadBalancerDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*godo.Client)

	log.Printf("[INFO] Deleting Load Balancer: %s", d.Id())
	_, err := client.LoadBalancers.Delete(d.Id())
	if err != nil {
		return fmt.Errorf("Error deleting Load Balancer: %s", err)
	}

	d.SetId("")
	return nil
}

func resourceDigitalOceanLoadBalancerImport(rs *schema.ResourceData, v interface{}) ([]*schema.ResourceData, error) {
	client := v.(*godo.Client)
	lb, _, err := client.LoadBalancers.Get(rs.Id())
	if err != nil {
		return nil, err
	}

	rs.Set("id", lb.ID)
	rs.Set("name", lb.Name)
	rs.Set("ip", lb.IP)
	rs.Set("algorithm", lb.Algorithm)
	rs.Set("region", lb.Region.Slug)

	dids := make([]interface{}, 0, len(lb.DropletIDs))
	for _, did := range lb.DropletIDs {
		dids = append(dids, did)
	}
	rs.Set("droplet_ids", schema.NewSet(
		func(dropletID interface{}) int { return dropletID.(int) },
		dids,
	))

	return []*schema.ResourceData{rs}, nil
}
