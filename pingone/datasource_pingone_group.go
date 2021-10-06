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

func datasourceGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceGroupRead,

		Schema: map[string]*schema.Schema{
			"environment_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"group_id": {
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
			"population_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_filter": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"external_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func datasourceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api_client := meta.(*pingone.APIClient)
	var diags diag.Diagnostics

	envID := d.Get("environment_id").(string)
	groupID := d.Get("group_id").(string)
	groupName := d.Get("name").(string)

	var resp pingone.Group
	if groupName != "" {

		filter := fmt.Sprintf("name eq \"%s\"", groupName)
		limit := int32(1)

		respList, r, err := api_client.ManagementAPIsGroupsApi.ReadAllGroups(context.Background(), envID).Filter(filter).Limit(limit).Execute()
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Error when calling `ManagementAPIsGroupsApi.ReadAllGroups``: %v", err),
				Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
			})

			return diags
		}

		resp = respList.Embedded.GetGroups()[0]

	} else {

		resp, r, err := api_client.ManagementAPIsGroupsApi.ReadOneGroup(context.Background(), envID, groupID).Execute()
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Error when calling `ManagementAPIsGroupsApi.ReadOneGroup``: %v", err),
				Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
			})

			return diags
		}
		log.Printf("Group found %s", resp.GetName())
	}

	d.SetId(resp.GetId())
	d.Set("group_id", resp.GetId())
	d.Set("name", resp.GetName())
	d.Set("description", resp.GetDescription())
	d.Set("population_id", resp.GetPopulation().Id)
	d.Set("user_filter", resp.GetUserFilter())
	d.Set("external_id", resp.GetExternalId())

	return diags
}
