package pingone

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/patrickcping/pingone-go"
)

func resourceApplicationAttributeMapping() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceApplicationAttributeMappingCreate,
		ReadContext:   resourceApplicationAttributeMappingRead,
		UpdateContext: resourceApplicationAttributeMappingUpdate,
		DeleteContext: resourceApplicationAttributeMappingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceApplicationAttributeMappingImport,
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
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"required": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"value": {
				Type:     schema.TypeString,
				Required: true,
			},
			"mapping_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceApplicationAttributeMappingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	envID := d.Get("environment_id").(string)
	appID := d.Get("application_id").(string)

	applicationAttributeMapping, err := expandApplicationAttributeMapping(d)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Cannot expand Application Attribute into SDK object",
			Detail:   fmt.Sprintf("Full error: %v\n", err),
		})

		return diags
	}

	log.Printf("[INFO] Creating PingOne Application Attribute: name %s", applicationAttributeMapping.GetName())

	resp, r, err := api_client.ManagementAPIsApplicationsApplicationAttributeMappingApi.CreateApplicationAttributeMapping(ctx, envID, appID).ApplicationAttributeMapping(applicationAttributeMapping).Execute()
	if (err != nil) && (r.StatusCode != 201) {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsApplicationsApplicationAttributeMappingApi.CreateApplicationAttributeMapping``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	d.SetId(resp.GetId())

	return resourceApplicationAttributeMappingRead(ctx, d, meta)
}

func resourceApplicationAttributeMappingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	attrMappingID := d.Id()
	envID := d.Get("environment_id").(string)
	appID := d.Get("application_id").(string)

	resp, r, err := api_client.ManagementAPIsApplicationsApplicationAttributeMappingApi.ReadOneApplicationAttributeMapping(ctx, envID, appID, attrMappingID).Execute()
	if err != nil {

		if r.StatusCode == 404 {
			log.Printf("[INFO] PingOne Application Mapping %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsApplicationsApplicationAttributeMappingApi.ReadOneApplicationAttributeMapping``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	d.Set("name", resp.GetName())
	d.Set("required", resp.GetRequired())
	d.Set("value", resp.GetValue())
	d.Set("mapping_type", resp.GetMappingType())

	return diags
}

func resourceApplicationAttributeMappingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	attrMappingID := d.Id()

	envID := d.Get("environment_id").(string)
	appID := d.Get("application_id").(string)

	applicationAttributeMapping, err := expandApplicationAttributeMapping(d)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Cannot expand Application Attribute into SDK object",
			Detail:   fmt.Sprintf("Full error: %v\n", err),
		})

		return diags
	}

	log.Printf("[INFO] Updating PingOne Application Attribute: name %s", applicationAttributeMapping.GetName())

	_, r, err := api_client.ManagementAPIsApplicationsApplicationAttributeMappingApi.UpdateApplicationAttributeMapping(ctx, envID, appID, attrMappingID).ApplicationAttributeMapping(applicationAttributeMapping).Execute()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsApplicationsApplicationAttributeMappingApi.UpdateApplicationAttributeMapping``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	return resourceApplicationAttributeMappingRead(ctx, d, meta)
}

func resourceApplicationAttributeMappingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	envID := d.Get("environment_id").(string)
	appID := d.Get("application_id").(string)

	attrMappingID := d.Id()

	_, err := api_client.ManagementAPIsApplicationsApplicationAttributeMappingApi.DeleteApplicationAttributeMapping(ctx, envID, appID, attrMappingID).Execute()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsApplicationsApplicationAttributeMappingApi.DeleteApplicationAttributeMapping``: %v", err),
		})

		return diags
	}

	return nil
}

func resourceApplicationAttributeMappingImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	attributes := strings.SplitN(d.Id(), "/", 3)

	if len(attributes) != 2 {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"envID/appID/attributeMappingID\"", d.Id())
	}

	envID, appID, attrMappingID := attributes[0], attributes[1], attributes[2]

	d.Set("environment_id", envID)
	d.Set("application_id", appID)
	d.SetId(attrMappingID)

	resourceApplicationAttributeMappingRead(ctx, d, meta)

	return []*schema.ResourceData{d}, nil
}

func expandApplicationAttributeMapping(d *schema.ResourceData) (pingone.ApplicationAttributeMapping, error) {

	applicationAttributeMapping := *pingone.NewApplicationAttributeMapping(d.Get("name").(string), d.Get("required").(bool), d.Get("value").(string))

	return applicationAttributeMapping, nil
}
