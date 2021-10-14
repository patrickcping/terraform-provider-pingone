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

func resourceGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGroupCreate,
		ReadContext:   resourceGroupRead,
		UpdateContext: resourceGroupUpdate,
		DeleteContext: resourceGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceGroupImport,
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
			"population_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"user_filter": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"external_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	envID := d.Get("environment_id").(string)
	groupName := d.Get("name").(string)
	groupDescription := d.Get("description").(string)
	populationID := d.Get("population_id").(string)
	userFilter := d.Get("user_filter").(string)
	externalID := d.Get("external_id").(string)

	log.Printf("[INFO] Creating PingOne Group: name %s", groupName)

	groupPopulation := *pingone.NewGroupPopulation(populationID) // NewGroupPopulation |  (optional)

	group := *pingone.NewGroup(groupName) // Group |  (optional)

	if populationID != "" {
		group.SetPopulation(groupPopulation)
	}

	if userFilter != "" {
		group.SetUserFilter(userFilter)
	}

	if groupDescription != "" {
		group.SetDescription(groupDescription)
	}

	if externalID != "" {
		group.SetExternalId(externalID)
	}

	log.Printf("Error when calling `ManagementAPIsGroupsApi.CreateGroup``: %v", group)
	resp, r, err := api_client.ManagementAPIsGroupsApi.CreateGroup(ctx, envID).Group(group).Execute()
	if (err != nil) && (r.StatusCode != 201) {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsGroupsApi.CreateGroup``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	d.SetId(resp.GetId())

	return resourceGroupRead(ctx, d, meta)
}

func resourceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	groupID := d.Id()
	envID := d.Get("environment_id").(string)

	resp, r, err := api_client.ManagementAPIsGroupsApi.ReadOneGroup(ctx, envID, groupID).Execute()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsGroupsApi.ReadOneGroup``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	d.Set("name", resp.GetName())
	d.Set("description", resp.GetDescription())
	d.Set("population_id", resp.GetPopulation().Id)
	d.Set("user_filter", resp.GetUserFilter())
	d.Set("external_id", resp.GetExternalId())

	return diags
}

func resourceGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	envID := d.Get("environment_id").(string)

	groupID := d.Id()
	groupName := d.Get("name").(string)
	groupDescription := d.Get("description").(string)
	populationID := d.Get("population_id").(string)
	userFilter := d.Get("user_filter").(string)
	externalID := d.Get("external_id").(string)

	log.Printf("[INFO] Updating PingOne Group: name %s", groupName)

	groupPopulation := *pingone.NewGroupPopulation(populationID) // NewGroupPopulation |  (optional)

	group := *pingone.NewGroup(groupName) // Group |  (optional)

	if populationID != "" {
		group.SetPopulation(groupPopulation)
	}

	if userFilter != "" {
		group.SetUserFilter(userFilter)
	}

	if groupDescription != "" {
		group.SetDescription(groupDescription)
	}

	if externalID != "" {
		group.SetExternalId(externalID)
	}

	_, r, err := api_client.ManagementAPIsGroupsApi.UpdateGroup(ctx, envID, groupID).Group(group).Execute()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsGroupsApi.UpdateGroup``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	return resourceGroupRead(ctx, d, meta)
}

func resourceGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	envID := d.Get("environment_id").(string)

	groupID := d.Id()

	_, err := api_client.ManagementAPIsGroupsApi.DeleteGroup(ctx, envID, groupID).Execute()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsGroupsApi.DeleteGroup``: %v", err),
		})

		return diags
	}

	return nil
}

func resourceGroupImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	attributes := strings.SplitN(d.Id(), "/", 2)

	if len(attributes) != 2 {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"envID/groupID\"", d.Id())
	}

	envID, groupID := attributes[0], attributes[1]

	d.Set("environment_id", envID)
	d.SetId(groupID)

	resourceGroupRead(ctx, d, meta)

	return []*schema.ResourceData{d}, nil
}
