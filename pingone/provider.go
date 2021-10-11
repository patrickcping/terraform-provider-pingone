package pingone

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// Provider -
func Provider() *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: descriptions["client_id"],
			},
			"client_secret": {
				Type:        schema.TypeString,
				Required:    true,
				Description: descriptions["client_secret"],
			},
			"environment_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: descriptions["environment_id"],
			},
			"region": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  descriptions["region"],
				ValidateFunc: validation.StringInSlice([]string{"EU", "US", "ASIA", "CA"}, false),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"pingone_environment":      resourceEnvironment(),
			"pingone_group":            resourceGroup(),
			"pingone_population":       resourcePopulation(),
			"pingone_role_assignment":  resourceUserRoleAssignment(),
			"pingone_schema_attribute": resourceSchemaAttribute(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"pingone_environment":  datasourceEnvironment(),
			"pingone_environments": datasourceEnvironments(),
			"pingone_group":        datasourceGroup(),
			"pingone_role":         datasourceRole(),
			"pingone_schema":       datasourceSchema(),
		},
		ConfigureContextFunc: providerConfigure,
	}

	return provider
}

var descriptions map[string]string

func init() {
	descriptions = map[string]string{
		"client_id":      "Client ID for the worker app client",
		"client_secret":  "Client secret for the worker app client",
		"environment_id": "Environment ID for the worker app client",
		"region":         "The PingOne region to use.  Options are EU, US, ASIA, CA",
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	log.Printf("[INFO] PingOne Client configuring")
	var diags diag.Diagnostics

	config := &p1ClientConfig{
		ClientId:      d.Get("client_id").(string),
		ClientSecret:  d.Get("client_secret").(string),
		EnvironmentID: d.Get("environment_id").(string),
		Region:        d.Get("region").(string),
	}

	client, err := config.ApiClient(ctx)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Error when getting access token",
			Detail:   fmt.Sprintf("Error when getting access token`: %v", err),
		})

		return nil, diags
	}

	return client, nil
}
