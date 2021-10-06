package pingone

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider -
func Provider() *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: descriptions["client_id"],
			},
			"client_secret": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: descriptions["client_secret"],
			},
			"access_token": {
				Type:        schema.TypeString,
				Required:    true,
				Description: descriptions["access_token"],
			},
			"environment_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: descriptions["environment_id"],
			},
			"domain_suffix": {
				Type:        schema.TypeString,
				Required:    true,
				Description: descriptions["domain_suffix"],
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"PINGONE_DOMAIN_SUFFIX"}, "eu"),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"pingone_environment":     resourceEnvironment(),
			"pingone_population":      resourcePopulation(),
			"pingone_role_assignment": resourceUserRoleAssignment(),
			"pingone_group":           resourceGroup(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"pingone_environment":  datasourceEnvironment(),
			"pingone_environments": datasourceEnvironments(),
			"pingone_role":         datasourceRole(),
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
		"access_token":   "An access token in lieu of the client ID and secret",
		"environment_id": "Environment ID for the worker app client",
		"domain_suffix":  "The domain suffix for the auth hostname and api hostname.  Value of eu translates to auth.pingone.eu and api.pingone.eu.  See https://apidocs.pingidentity.com/pingone/platform/v1/api/#top",
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	log.Printf("[INFO] PingOne Client configuring")

	config := &p1ClientConfig{
		ClientId:      d.Get("client_id").(string),
		ClientSecret:  d.Get("client_secret").(string),
		AccessToken:   d.Get("access_token").(string),
		EnvironmentID: d.Get("environment_id").(string),
		DomainSuffix:  d.Get("domain_suffix").(string),
	}

	client, _ := config.ApiClient()

	return client, nil
}
