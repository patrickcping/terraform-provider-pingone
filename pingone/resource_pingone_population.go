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

func resourcePopulation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePopulationCreate,
		ReadContext:   resourcePopulationRead,
		UpdateContext: resourcePopulationUpdate,
		DeleteContext: resourcePopulationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourcePopulationImport,
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
		},
	}
}

func resourcePopulationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	envID := d.Get("environment_id").(string)
	popName := d.Get("name").(string)
	popDescription := d.Get("description").(string)

	log.Printf("[INFO] Creating PingOne Population: name %s", popName)

	population := *pingone.NewPopulation(popName) // Population |  (optional)
	population.SetDescription(popDescription)

	resp, r, err := api_client.ManagementAPIsPopulationsApi.CreatePopulation(ctx, envID).Population(population).Execute()
	if (err != nil) && (r.StatusCode != 201) {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsPopulationsApi.CreatePopulation``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	d.SetId(resp.GetId())

	return resourcePopulationRead(ctx, d, meta)
}

func resourcePopulationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	popID := d.Id()
	envID := d.Get("environment_id").(string)

	resp, r, err := api_client.ManagementAPIsPopulationsApi.ReadOnePopulation(ctx, envID, popID).Execute()
	if err != nil {

		if r.StatusCode == 404 {
			log.Printf("[INFO] PingOne Population %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsPopulationsApi.ReadOnePopulation``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	d.Set("name", resp.GetName())
	d.Set("description", resp.GetDescription())

	return diags
}

func resourcePopulationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	envID := d.Get("environment_id").(string)

	popID := d.Id()
	popName := d.Get("name").(string)
	popDescription := d.Get("description").(string)

	population := *pingone.NewPopulation(popName) // Population |  (optional)
	population.SetDescription(popDescription)

	_, r, err := api_client.ManagementAPIsPopulationsApi.UpdatePopulation(ctx, envID, popID).Population(population).Execute()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsPopulationsApi.UpdatePopulation``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	return resourcePopulationRead(ctx, d, meta)
}

func resourcePopulationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	envID := d.Get("environment_id").(string)

	popID := d.Id()

	_, err := api_client.ManagementAPIsPopulationsApi.DeletePopulation(ctx, envID, popID).Execute()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsPopulationsApi.DeletePopulation``: %v", err),
		})

		return diags
	}

	return nil
}

func resourcePopulationImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	attributes := strings.SplitN(d.Id(), "/", 2)

	if len(attributes) != 2 {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"envID/populationID\"", d.Id())
	}

	envID, populationID := attributes[0], attributes[1]

	d.Set("environment_id", envID)
	d.SetId(populationID)

	resourcePopulationRead(ctx, d, meta)

	return []*schema.ResourceData{d}, nil
}
