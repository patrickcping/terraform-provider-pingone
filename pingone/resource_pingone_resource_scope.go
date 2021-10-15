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

func resourceResourceScope() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceResourceScopeCreate,
		ReadContext:   resourceResourceScopeRead,
		UpdateContext: resourceResourceScopeUpdate,
		DeleteContext: resourceResourceScopeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceResourceScopeImport,
		},

		Schema: map[string]*schema.Schema{
			"environment_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"resource_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"schema_attributes": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceResourceScopeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	envID := d.Get("environment_id").(string)
	resourceID := d.Get("resource_id").(string)

	resourceScope, err := expandResourceScope(d)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Cannot expand Resource Scope into SDK object",
			Detail:   fmt.Sprintf("Full error: %v\n", err),
		})

		return diags
	}

	log.Printf("[INFO] Creating PingOne Resource Scope: name %s", resourceScope.GetName())

	resp, r, err := api_client.ManagementAPIsResourcesResourceScopesApi.CreateResourceScope(ctx, envID, resourceID).ResourceScope(resourceScope).Execute()
	if (err != nil) && (r.StatusCode != 201) {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsResourcesResourceScopesApi.CreateResourceScope``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	d.SetId(resp.GetId())

	return resourceResourceScopeRead(ctx, d, meta)
}

func resourceResourceScopeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	resourceScopeID := d.Id()
	envID := d.Get("environment_id").(string)
	resourceID := d.Get("resource_id").(string)

	resp, r, err := api_client.ManagementAPIsResourcesResourceScopesApi.ReadOneResourceScope(ctx, envID, resourceID, resourceScopeID).Execute()
	if err != nil {

		if r.StatusCode == 404 {
			log.Printf("[INFO] PingOne Resource Scope %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsResourcesResourceScopesApi.ReadOneResourceScope``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	d.Set("name", resp.GetName())
	d.Set("description", resp.GetDescription())
	d.Set("schema_attributes", resp.GetSchemaAttributes())

	return diags
}

func resourceResourceScopeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	envID := d.Get("environment_id").(string)
	resourceID := d.Get("resource_id").(string)

	resourceScopeID := d.Id()

	resourceScope, err := expandResourceScope(d)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Cannot expand Resource Scope into SDK object",
			Detail:   fmt.Sprintf("Full error: %v\n", err),
		})

		return diags
	}

	log.Printf("[INFO] Updating PingOne Resource Scope: name %s", resourceScope.GetName())

	_, r, err := api_client.ManagementAPIsResourcesResourceScopesApi.UpdateResourceScope(ctx, envID, resourceID, resourceScopeID).ResourceScope(resourceScope).Execute()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsResourcesResourceScopesApi.UpdateResourceScope``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	return resourceResourceScopeRead(ctx, d, meta)
}

func resourceResourceScopeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	envID := d.Get("environment_id").(string)
	resourceID := d.Get("resource_id").(string)

	resourceScopeID := d.Id()

	_, err := api_client.ManagementAPIsResourcesResourceScopesApi.DeleteResourceScope(ctx, envID, resourceID, resourceScopeID).Execute()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsResourcesResourceScopesApi.DeleteResourceScope``: %v", err),
		})

		return diags
	}

	return nil
}

func resourceResourceScopeImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	attributes := strings.SplitN(d.Id(), "/", 2)

	if len(attributes) != 2 {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"envID/resourceID/resourceScopeID\"", d.Id())
	}

	envID, resourceID, resourceScopeID := attributes[0], attributes[1], attributes[2]

	d.Set("environment_id", envID)
	d.Set("resource_id", resourceID)
	d.SetId(resourceScopeID)

	resourceResourceScopeRead(ctx, d, meta)

	return []*schema.ResourceData{d}, nil
}

func expandResourceScope(d *schema.ResourceData) (pingone.ResourceScope, error) {

	resource := *pingone.NewResourceScope(d.Get("name").(string))

	if v, ok := d.GetOk("description"); ok {
		resource.SetDescription(v.(string))
	}

	if v, ok := d.GetOk("schema_attributes"); ok {
		resource.SetSchemaAttributes(v.([]string))
	}

	return resource, nil
}
