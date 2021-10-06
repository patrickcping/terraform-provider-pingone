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

func datasourceEnvironments() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceEnvironmentsRead,

		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"environments": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
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
				},
			},
		},
	}
}

func datasourceEnvironmentsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api_client := meta.(*pingone.APIClient)
	var diags diag.Diagnostics

	filter := d.Get("filter").(string)

	var resp []pingone.Environment

	limit := int32(1000)
	respList, r, err := api_client.ManagementAPIsEnvironmentsApi.ReadAllEnvironments(context.Background()).Limit(limit).Filter(filter).Execute()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsEnvironmentsApi.ReadAllEnvironments``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	resp = respList.Embedded.GetEnvironments()
	log.Printf("Environments found")

	environments := make([]interface{}, 0, len(resp))
	for _, environment := range resp {

		environments = append(environments, map[string]interface{}{
			"environment_id": environment.GetId(),
			"name":           environment.GetName(),
			"description":    environment.GetDescription(),
			"type":           environment.GetType(),
			"region":         environment.GetRegion(),
			"license_id":     environment.License.GetId(),
		})
	}

	d.Set("environments", environments)

	return diags
}
