package pingone

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/patrickcping/pingone-go"
)

func resourceApplicationResourceGrant() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceApplicationResourceGrantCreate,
		ReadContext:   resourceApplicationResourceGrantRead,
		UpdateContext: resourceApplicationResourceGrantUpdate,
		DeleteContext: resourceApplicationResourceGrantDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceApplicationResourceGrantImport,
		},

		Schema: map[string]*schema.Schema{
			"environment_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"application_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
	}
}

func resourceApplicationResourceGrantCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	envID := d.Get("environment_id").(string)
	appID := d.Get("application_id").(string)

	applicationResourceGrant, err := expandApplicationResourceGrant(d)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Cannot expand Application into SDK object",
			Detail:   fmt.Sprintf("Full error: %v\n", err),
		})

		return diags
	}

	log.Printf("[INFO] Creating PingOne Application resource grant")

	resp, r, err := api_client.ManagementAPIsApplicationsApplicationResourceGrantsApi.CreateApplicationGrant(ctx, envID, appID).ApplicationResourceGrant(applicationResourceGrant).Execute()
	if (err != nil) || (r.StatusCode != 201) {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsApplicationsApplicationResourceGrantsApi.CreateGrant``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	d.SetId(resp.GetId())

	return resourceApplicationResourceGrantRead(ctx, d, meta)
}

func resourceApplicationResourceGrantRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	grantID := d.Id()
	envID := d.Get("environment_id").(string)
	appID := d.Get("application_id").(string)

	resp, r, err := api_client.ManagementAPIsApplicationsApplicationResourceGrantsApi.ReadOneApplicationGrant(ctx, envID, appID, grantID).Execute()
	if err != nil {

		if r.StatusCode == 404 {
			log.Printf("[INFO] PingOne Application Resource Grant %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsApplicationsApplicationResourceGrantsApi.ReadOneApplicationGrant``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	d.Set("resource", resp.Resource.GetId())
	d.Set("scopes", flattenAppResourceGrantScopes(resp.GetScopes()))

	return diags
}

func resourceApplicationResourceGrantUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	grantID := d.Id()

	envID := d.Get("environment_id").(string)
	appID := d.Get("application_id").(string)

	applicationResourceGrant, err := expandApplicationResourceGrant(d)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Cannot expand Application into SDK object",
			Detail:   fmt.Sprintf("Full error: %v\n", err),
		})

		return diags
	}

	log.Printf("[INFO] Creating PingOne Application resource grant")

	_, r, err := api_client.ManagementAPIsApplicationsApplicationResourceGrantsApi.UpdateApplicationGrant(ctx, envID, appID, grantID).ApplicationResourceGrant(applicationResourceGrant).Execute()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsApplicationsApplicationResourceGrantsApi.UpdateApplicationGrant``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	return resourceApplicationResourceGrantRead(ctx, d, meta)
}

func resourceApplicationResourceGrantDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	envID := d.Get("environment_id").(string)
	appID := d.Get("application_id").(string)

	grantID := d.Id()

	_, err := api_client.ManagementAPIsApplicationsApplicationResourceGrantsApi.DeleteApplicationGrant(ctx, envID, appID, grantID).Execute()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsApplicationsApplicationResourceGrantsApi.DeleteApplicationGrant``: %v", err),
		})

		return diags
	}

	return nil
}

func resourceApplicationResourceGrantImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	attributes := strings.SplitN(d.Id(), "/", 3)

	if len(attributes) != 2 {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"envID/applicationID/grantID\"", d.Id())
	}

	envID, appID, grantID := attributes[0], attributes[1], attributes[2]

	d.Set("environment_id", envID)
	d.Set("application_id", appID)
	d.SetId(grantID)

	resourceApplicationResourceGrantRead(ctx, d, meta)

	return []*schema.ResourceData{d}, nil
}

func expandApplicationResourceGrant(d *schema.ResourceData) (pingone.ApplicationResourceGrant, error) {

	scopesIn := d.Get("scopes").([]interface{})
	scopes := make([]pingone.ApplicationResourceGrantScopes, 0, len(scopesIn))
	for _, scope := range scopesIn {
		scopes = append(scopes, pingone.ApplicationResourceGrantScopes{
			Id: scope.(string),
		})
	}

	sort.Slice(scopes, func(i, j int) bool {
		return scopes[i].GetId() < scopes[j].GetId()
	})

	resource := *pingone.NewApplicationResourceGrantResource(d.Get("resource_id").(string))

	applicationResourceGrant := *pingone.NewApplicationResourceGrant(resource, scopes)

	return applicationResourceGrant, nil
}

func flattenAppResourceGrantScopes(in []pingone.ApplicationResourceGrantScopes) []string {

	items := make([]string, 0, len(in))
	for _, v := range in {

		items = append(items, v.GetId())
	}

	sort.Strings(items)
	return items
}
