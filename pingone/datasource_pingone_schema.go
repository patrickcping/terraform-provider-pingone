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

func datasourceSchema() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceSchemaRead,

		Schema: map[string]*schema.Schema{
			"environment_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"schema_id": {
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
		},
	}
}

func datasourceSchemaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	envID := d.Get("environment_id").(string)
	schemaID := d.Get("schema_id").(string)
	schemaName := d.Get("name").(string)

	var resp pingone.Schema
	if schemaName != "" {

		respList, r, err := api_client.ManagementAPIsSchemasApi.ReadAllSchemas(ctx, envID).Execute()
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Error when calling `ManagementAPIsSchemasApi.ReadAllSchemas``: %v", err),
				Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
			})

			return diags
		}

		for _, v := range respList.Embedded.GetSchemas() {
			if v.GetName() == schemaName {
				resp = v
				log.Printf("Schema found %s", resp.GetName())
			}
		}

	} else {

		resp, r, err := api_client.ManagementAPIsSchemasApi.ReadOneSchema(ctx, envID, schemaID).Execute()
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Error when calling `ManagementAPIsSchemasApi.ReadOneSchema``: %v", err),
				Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
			})

			return diags
		}
		log.Printf("Schema found %s", resp.GetName())
	}

	d.SetId(resp.GetId())
	d.Set("schema_id", resp.GetId())
	d.Set("name", resp.GetName())
	d.Set("description", resp.GetDescription())

	return diags
}
