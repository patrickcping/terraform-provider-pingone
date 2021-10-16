package pingone

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/patrickcping/pingone-go"
)

func datasourceApplicationResourceGrants() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceApplicationResourceGrantsRead,

		Schema: map[string]*schema.Schema{
			"environment_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"application_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"resource_grants": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"resource_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"scopes": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func datasourceApplicationResourceGrantsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	envID := d.Get("environment_id").(string)
	appID := d.Get("application_id").(string)

	var resp []pingone.ApplicationResourceGrant

	respList, r, err := api_client.ManagementAPIsApplicationsApplicationResourceGrantsApi.ReadAllApplicationGrants(ctx, envID, appID).Execute()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsApplicationsApplicationResourceGrantsApi.ReadAllApplicationGrants``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	resp = respList.Embedded.GetGrants()
	log.Printf("Application Resource Grants found")

	grants := make([]interface{}, 0, len(resp))
	for _, grant := range resp {

		grants = append(grants, map[string]interface{}{
			"id":          grant.GetId(),
			"resource_id": grant.Resource.GetId(),
			"scopes":      flattenAppResourceGrantScopes(grant.GetScopes()),
		})
	}

	d.Set("resource_grants", grants)

	return diags
}
