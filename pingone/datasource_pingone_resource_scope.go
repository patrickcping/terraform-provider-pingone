package pingone

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/patrickcping/pingone-go"
)

func datasourceResourceScope() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceResourceScopeRead,

		Schema: map[string]*schema.Schema{
			"environment_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"resource_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"scope_id": {
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
			"schema_attributes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func datasourceResourceScopeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	envID := d.Get("environment_id").(string)
	resourceID := d.Get("resource_id").(string)
	scopeID := d.Get("scope_id").(string)
	scopeName := d.Get("name").(string)

	var resp pingone.ResourceScope
	if scopeName != "" {

		respList, r, err := api_client.ManagementAPIsResourcesResourceScopesApi.ReadAllResourceScopes(ctx, envID, resourceID).Execute()
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Error when calling `ManagementAPIsResourcesResourceScopesApi.ReadAllResourceScopes``: %v", err),
				Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
			})

			return diags
		}

		for _, v := range respList.Embedded.GetScopes() {
			if v.GetName() == scopeName {
				resp = v
				log.Printf("Resource Scope found %s", resp.GetName())
			}
		}

	} else {

		resp, r, err := api_client.ManagementAPIsResourcesResourceScopesApi.ReadOneResourceScope(ctx, envID, resourceID, scopeID).Execute()
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Error when calling `ManagementAPIsResourcesResourceScopesApi.ReadOneResourceScope``: %v", err),
				Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
			})

			return diags
		}
		log.Printf("ResourceScope found %s", resp.GetName())
	}

	d.SetId(resp.GetId())
	d.Set("scope_id", resp.GetId())
	d.Set("name", resp.GetName())
	d.Set("description", resp.GetDescription())
	d.Set("schema_attributes", resp.GetSchemaAttributes())

	return diags
}
