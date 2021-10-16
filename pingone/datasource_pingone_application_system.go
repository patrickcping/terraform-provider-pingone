package pingone

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/patrickcping/pingone-go"
)

func datasourceApplicationSystem() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceApplicationSystemRead,

		Schema: map[string]*schema.Schema{
			"environment_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"PING_ONE_ADMIN_CONSOLE", "PING_ONE_PORTAL", "PING_ONE_SELF_SERVICE"}, false),
			},
			"application_id": {
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
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"home_page_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"access_control": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"role_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"group": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"groups": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func datasourceApplicationSystemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	envID := d.Get("environment_id").(string)
	appType := d.Get("type").(string)

	var resp pingone.ApplicationOIDC

	respList, r, err := api_client.ManagementAPIsApplicationsApplicationsApi.ReadAllApplications(ctx, envID).Execute()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsApplicationsApplicationsApi.ReadAllApplications``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	for _, v := range respList.Embedded.GetApplications() {

		b, err := json.Marshal(v)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Cannot marshal application json to byte",
				Detail:   fmt.Sprintf("Full response: %v\n", err),
			})

			return diags
		}

		application := pingone.ApplicationOIDC{}
		json.Unmarshal([]byte(b), &application)

		if appType != "" && application.GetType() == appType {
			resp = application
			log.Printf("Application found %s", resp.GetName())
			break
		}

	}

	if id, ok := resp.GetIdOk(); ok {

		d.SetId(*id)
		d.Set("application_id", id)
		d.Set("name", resp.GetName())
		d.Set("description", resp.GetDescription())
		d.Set("type", resp.GetType())
		d.Set("enabled", resp.GetEnabled())
		d.Set("home_page_url", resp.GetHomePageUrl())

		if v, ok := resp.GetAccessControlOk(); ok {
			accessControlFlattened, err := flattenApplicationAccessControl(v)

			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Cannot flatten Application from SDK object",
					Detail:   fmt.Sprintf("Full error: %v\n", err),
				})

				return diags
			}
			d.Set("access_control", accessControlFlattened)
		}
	} else {
		d.SetId("")
	}

	return diags
}
