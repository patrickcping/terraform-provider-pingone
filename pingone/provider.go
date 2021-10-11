package pingone

import (
	"context"
	"log"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/patrickcping/pingone-go"
)

var stderr = os.Stderr

func New() tfsdk.Provider {
	return &provider{}
}

type provider struct {
	configured bool
	client     *pingone.APIClient
}

// GetSchema -
func (p *provider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"client_id": {
				Type:        types.StringType,
				Optional:    true,
				Description: descriptions["client_id"],
			},
			"client_secret": {
				Type:        types.StringType,
				Optional:    true,
				Sensitive:   true,
				Description: descriptions["client_secret"],
			},
			"access_token": {
				Type:        types.StringType,
				Required:    true,
				Description: descriptions["access_token"],
			},
			"environment_id": {
				Type:        types.StringType,
				Required:    true,
				Description: descriptions["environment_id"],
			},
			"domain_suffix": {
				Type:        types.StringType,
				Required:    true,
				Description: descriptions["domain_suffix"],
			},
		},
	}, nil
}

var descriptions map[string]string

func (p *provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {
	log.Printf("[INFO] PingOne Client configuring")

	var config providerData

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, _ := config.ApiClient()

	p.client = client
	p.configured = true

}

// GetResources - Defines provider resources
func (p *provider) GetResources(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{
		"pingone_environment": resourceEnvironmentType{},
		// "pingone_group":            resourceGroupType{},
		// "pingone_population":       resourcePopulationType{},
		// "pingone_role_assignment":  resourceUserRoleAssignmentType{},
		// "pingone_schema_attribute": resourceSchemaAttributeType{},
	}, nil
}

// GetDataSources - Defines provider data sources
func (p *provider) GetDataSources(_ context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{
		// "pingone_environment":  datasourceEnvironmentType{},
		// "pingone_environments": datasourceEnvironmentsType{},
		// "pingone_group":        datasourceGroupType{},
		// "pingone_role":         datasourceRoleType{},
		// "pingone_schema":       datasourceSchemaType{},
	}, nil
}
