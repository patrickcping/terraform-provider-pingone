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

func resourceResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceResourceCreate,
		ReadContext:   resourceResourceRead,
		UpdateContext: resourceResourceUpdate,
		DeleteContext: resourceResourceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceResourceImport,
		},

		Schema: map[string]*schema.Schema{
			"environment_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"audience": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"access_token_validity_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  3600,
			},
		},
	}
}

func resourceResourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	envID := d.Get("environment_id").(string)

	resource, err := expandResource(d)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Cannot expand Resource into SDK object",
			Detail:   fmt.Sprintf("Full error: %v\n", err),
		})

		return diags
	}

	log.Printf("[INFO] Creating PingOne Resource: name %s", resource.GetName())

	resp, r, err := api_client.ManagementAPIsResourcesResourcesApi.CreateResource(ctx, envID).Resource(resource).Execute()
	if (err != nil) && (r.StatusCode != 201) {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsResourcesResourcesApi.CreateResource``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	d.SetId(resp.GetId())

	return resourceResourceRead(ctx, d, meta)
}

func resourceResourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	resourceID := d.Id()
	envID := d.Get("environment_id").(string)

	resp, r, err := api_client.ManagementAPIsResourcesResourcesApi.ReadOneResource(ctx, envID, resourceID).Execute()
	if err != nil {

		if r.StatusCode == 404 {
			log.Printf("[INFO] PingOne Resource %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsResourcesResourcesApi.ReadOneResource``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	d.Set("name", resp.GetName())
	d.Set("description", resp.GetDescription())
	d.Set("type", resp.GetType())
	d.Set("audience", resp.GetAudience())
	d.Set("access_token_validity_seconds", resp.GetAccessTokenValiditySeconds())

	return diags
}

func resourceResourceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	envID := d.Get("environment_id").(string)

	resourceID := d.Id()

	resource, err := expandResource(d)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Cannot expand Resource into SDK object",
			Detail:   fmt.Sprintf("Full error: %v\n", err),
		})

		return diags
	}

	log.Printf("[INFO] Updating PingOne Resource: name %s", resource.GetName())

	_, r, err := api_client.ManagementAPIsResourcesResourcesApi.UpdateResource(ctx, envID, resourceID).Resource(resource).Execute()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsResourcesResourcesApi.UpdateResource``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	return resourceResourceRead(ctx, d, meta)
}

func resourceResourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	envID := d.Get("environment_id").(string)

	resourceID := d.Id()

	_, err := api_client.ManagementAPIsResourcesResourcesApi.DeleteResource(ctx, envID, resourceID).Execute()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsResourcesResourcesApi.DeleteResource``: %v", err),
		})

		return diags
	}

	return nil
}

func resourceResourceImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	attributes := strings.SplitN(d.Id(), "/", 2)

	if len(attributes) != 2 {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"envID/resourceID\"", d.Id())
	}

	envID, resourceID := attributes[0], attributes[1]

	d.Set("environment_id", envID)
	d.SetId(resourceID)

	resourceResourceRead(ctx, d, meta)

	return []*schema.ResourceData{d}, nil
}

func expandResource(d *schema.ResourceData) (pingone.Resource, error) {

	resource := *pingone.NewResource(d.Get("name").(string))
	resource.SetType("CUSTOM")

	if v, ok := d.GetOk("description"); ok {
		resource.SetDescription(v.(string))
	}

	if v, ok := d.GetOk("audience"); ok {
		resource.SetAudience(v.(string))
	}

	if v, ok := d.GetOk("access_token_validity_seconds"); ok {
		resource.SetAccessTokenValiditySeconds(int32(v.(int)))
	}

	return resource, nil
}
