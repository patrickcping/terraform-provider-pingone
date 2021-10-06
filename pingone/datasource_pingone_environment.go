package pingone

//https://learn.hashicorp.com/tutorials/terraform/provider-setup?in=terraform/providers
import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/patrickcping/pingone-go"
)

func datasourceEnvironment() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceEnvironmentRead,

		Schema: map[string]*schema.Schema{
			"environment_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"license_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func datasourceEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api_client := meta.(*pingone.APIClient)
	var diags diag.Diagnostics

	envID := d.Get("environment_id").(string)
	envName := d.Get("name").(string)

	var resp pingone.Environment
	if envName != "" {

		limit := int32(1000)
		filter := fmt.Sprintf("name sw \"%s\"", envName) // need the eq filter
		respList, r, err := api_client.ManagementAPIsEnvironmentsApi.ReadAllEnvironments(context.Background()).Limit(limit).Filter(filter).Execute()
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Error when calling `ManagementAPIsEnvironmentsApi.ReadAllEnvironments``: %v", err),
				Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
			})

			return diags
		}

		resp = respList.Embedded.GetEnvironments()[0]
		log.Printf("Environment found %s", *resp.Name)

	} else {

		resp, r, err := api_client.ManagementAPIsEnvironmentsApi.ReadOneEnvironment(context.Background(), envID).Execute()
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Error when calling `ManagementAPIsEnvironmentsApi.ReadOneEnvironment``: %v", err),
				Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
			})

			return diags
		}
		log.Printf("Environment found %s", *resp.Name)
	}

	d.SetId(resp.GetId())
	d.Set("environment_id", resp.GetId())
	d.Set("name", resp.GetName())
	d.Set("description", resp.GetDescription())
	d.Set("type", resp.GetType())
	d.Set("region", resp.GetRegion())
	d.Set("license_id", resp.GetLicense().Id)

	return diags
}
