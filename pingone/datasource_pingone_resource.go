package pingone

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/patrickcping/pingone-go"
)

func datasourceResource() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceResourceRead,

		Schema: map[string]*schema.Schema{
			"environment_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"resource_id": {
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
			"audience": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"access_token_validity_seconds": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func datasourceResourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	envID := d.Get("environment_id").(string)
	resourceID := d.Get("resource_id").(string)
	resourceName := d.Get("name").(string)

	var resp pingone.Resource
	if resourceName != "" {

		respList, r, err := api_client.ManagementAPIsResourcesResourcesApi.ReadAllResources(ctx, envID).Execute()
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Error when calling `ManagementAPIsResourcesResourcesApi.ReadAllResources``: %v", err),
				Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
			})

			return diags
		}

		for _, v := range respList.Embedded.GetResources() {
			if v.GetName() == resourceName {
				resp = v
				log.Printf("Resource found %s", resp.GetName())
			}
		}

	} else {

		resp, r, err := api_client.ManagementAPIsResourcesResourcesApi.ReadOneResource(ctx, envID, resourceID).Execute()
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Error when calling `ManagementAPIsResourcesResourcesApi.ReadOneResource``: %v", err),
				Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
			})

			return diags
		}
		log.Printf("Resource found %s", resp.GetName())
	}

	d.SetId(resp.GetId())
	d.Set("resource_id", resp.GetId())
	d.Set("name", resp.GetName())
	d.Set("description", resp.GetDescription())
	d.Set("type", resp.GetType())
	d.Set("audience", resp.GetAudience())
	d.Set("access_token_validity_seconds", resp.GetAccessTokenValiditySeconds())

	return diags
}
