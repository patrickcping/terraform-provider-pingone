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

func datasourceRole() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceRoleRead,

		Schema: map[string]*schema.Schema{
			"role_id": {
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

func datasourceRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerIndex, p1Client.regionUrlIndex)
	var diags diag.Diagnostics

	roleID := d.Get("role_id").(string)
	roleName := d.Get("name").(string)

	var resp pingone.Role
	if roleName != "" {

		respList, r, err := api_client.ManagementAPIsRolesApi.ReadAllRoles(ctx).Execute()
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Error when calling `ManagementAPIsRolesApi.ReadAllRoles``: %v", err),
				Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
			})

			return diags
		}

		for _, v := range respList.Embedded.GetRoles() {
			if v.GetName() == roleName {
				resp = v
				log.Printf("Role found %s", resp.GetName())
			}
		}

	} else {

		resp, r, err := api_client.ManagementAPIsRolesApi.ReadOneRole(ctx, roleID).Execute()
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Error when calling `ManagementAPIsRolesApi.ReadOneRole``: %v", err),
				Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
			})

			return diags
		}
		log.Printf("Role found %s", resp.GetName())
	}

	d.SetId(resp.GetId())
	d.Set("role_id", resp.GetId())
	d.Set("name", resp.GetName())
	d.Set("description", resp.GetDescription())

	return diags
}
