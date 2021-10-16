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

func resourceGatewayCredential() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGatewayCredentialCreate,
		ReadContext:   resourceGatewayCredentialRead,
		DeleteContext: resourceGatewayCredentialDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceGatewayCredentialImport,
		},

		Schema: map[string]*schema.Schema{
			"environment_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"credential": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"console_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"api_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auth_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceGatewayCredentialCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	envID := d.Get("environment_id").(string)
	gatewayID := d.Get("gateway_id").(string)

	log.Printf("[INFO] Creating PingOne Gateway Credential: gateway %s", gatewayID)

	resp, r, err := api_client.ManagementAPIsGatewayManagementGatewayCredentialsApi.CreateGatewayCredential(ctx, envID, gatewayID).Execute()
	if (err != nil) || (r.StatusCode != 201) {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsGatewayManagementGatewayCredentialsApi.CreateGatewayCredential``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	d.SetId(resp.GetId())
	d.Set("credential", resp.GetCredential())
	d.Set("console_url", resp.GetConsoleUrl())
	d.Set("api_url", resp.GetApiUrl())
	d.Set("auth_url", resp.GetAuthUrl())

	return resourceGatewayCredentialRead(ctx, d, meta)
}

func resourceGatewayCredentialRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	/// Do something here for sure
	/*


		if r.StatusCode == 404 {
			log.Printf("[INFO] PingOne Application Resource Grant %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
	*/

	return diags
}

func resourceGatewayCredentialDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	envID := d.Get("environment_id").(string)
	gatewayID := d.Get("gateway_id").(string)

	gatewayCredentialID := d.Id()

	_, err := api_client.ManagementAPIsGatewayManagementGatewayCredentialsApi.DeleteGatewayCredential(ctx, envID, gatewayID, gatewayCredentialID).Execute()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsGatewayManagementGatewayCredentialsApi.DeleteGatewayCredential``: %v", err),
		})

		return diags
	}

	return nil
}

func resourceGatewayCredentialImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	attributes := strings.SplitN(d.Id(), "/", 2)

	if len(attributes) != 2 {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"envID/gatewayID/gatewayCredentialID\"", d.Id())
	}

	envID, gatewayID, gatewayCredentialID := attributes[0], attributes[1], attributes[2]

	d.Set("environment_id", envID)
	d.Set("gateway_id", gatewayID)
	d.SetId(gatewayCredentialID)

	resourceGatewayCredentialRead(ctx, d, meta)

	return []*schema.ResourceData{d}, nil
}
