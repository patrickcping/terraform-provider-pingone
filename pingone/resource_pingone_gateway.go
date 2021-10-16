package pingone

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/patrickcping/pingone-go"
)

func resourceGateway() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGatewayCreate,
		ReadContext:   resourceGatewayRead,
		UpdateContext: resourceGatewayUpdate,
		DeleteContext: resourceGatewayDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceGatewayImport,
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
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"LDAP", "PING_FEDERATE", "PING_INTELLIGENCE"}, false),
			},
			"enabled": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceGatewayCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	envID := d.Get("environment_id").(string)

	gateway, err := expandGateway(d)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Cannot expand Application into SDK object",
			Detail:   fmt.Sprintf("Full error: %v\n", err),
		})

		return diags
	}

	log.Printf("[INFO] Creating PingOne Gateway: name %s", gateway.GetName())

	resp, r, err := api_client.ManagementAPIsGatewayManagementGatewaysApi.CreateGateway(ctx, envID).OneOfGatewayGatewayLDAP(gateway).Execute()
	if (err != nil) || (r.StatusCode != 201) {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsGatewayManagementGatewaysApi.CreateGateway``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	d.SetId(resp.(map[string]interface{})["id"].(string))

	return resourceGatewayRead(ctx, d, meta)
}

func resourceGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	gatewayID := d.Id()
	envID := d.Get("environment_id").(string)

	resp, r, err := api_client.ManagementAPIsGatewayManagementGatewaysApi.ReadOneGateway(ctx, envID, gatewayID).Execute()
	if err != nil {

		if r.StatusCode == 404 {
			log.Printf("[INFO] PingOne Gateway %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsGatewayManagementGatewaysApi.ReadOneGateway``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	b, err := json.Marshal(resp)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Cannot marshal gateway json to byte",
			Detail:   fmt.Sprintf("Full response: %v\n", err),
		})

		return diags
	}

	gateway := pingone.Gateway{}
	json.Unmarshal([]byte(b), &gateway)

	d.Set("name", gateway.GetName())
	d.Set("description", gateway.GetDescription())
	d.Set("type", gateway.GetType())
	d.Set("enabled", gateway.GetEnabled())

	return diags
}

func resourceGatewayUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	gatewayID := d.Id()
	envID := d.Get("environment_id").(string)

	gateway, err := expandGateway(d)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Cannot expand Application into SDK object",
			Detail:   fmt.Sprintf("Full error: %v\n", err),
		})

		return diags
	}

	log.Printf("[INFO] Updating PingOne Gateway: name %s", gateway.GetName())

	_, r, err := api_client.ManagementAPIsGatewayManagementGatewaysApi.UpdateGateway(ctx, envID, gatewayID).OneOfGatewayGatewayLDAP(gateway).Execute()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsGatewayManagementGatewaysApi.UpdateGateway``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	return resourceGatewayRead(ctx, d, meta)
}

func resourceGatewayDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	envID := d.Get("environment_id").(string)

	gatewayID := d.Id()

	_, err := api_client.ManagementAPIsGatewayManagementGatewaysApi.DeleteGateway(ctx, envID, gatewayID).Execute()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsGatewayManagementGatewaysApi.DeleteGateway``: %v", err),
		})

		return diags
	}

	return nil
}

func resourceGatewayImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	attributes := strings.SplitN(d.Id(), "/", 2)

	if len(attributes) != 2 {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"envID/gatewayID\"", d.Id())
	}

	envID, gatewayID := attributes[0], attributes[1]

	d.Set("environment_id", envID)
	d.SetId(gatewayID)

	resourceGatewayRead(ctx, d, meta)

	return []*schema.ResourceData{d}, nil
}

func expandGateway(d *schema.ResourceData) (pingone.Gateway, error) {

	gateway := *pingone.NewGateway(d.Get("name").(string), d.Get("type").(string), d.Get("enabled").(bool))
	if v, ok := d.GetOk("description"); ok {
		gateway.SetDescription(v.(string))
	}

	return gateway, nil
}
