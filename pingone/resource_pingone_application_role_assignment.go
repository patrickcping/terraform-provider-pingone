package pingone

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/patrickcping/pingone-go"
)

func resourceApplicationRoleAssignment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceApplicationRoleAssignmentCreate,
		ReadContext:   resourceApplicationRoleAssignmentRead,
		DeleteContext: resourceApplicationRoleAssignmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceApplicationRoleAssignmentImport,
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
			"role_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"scope_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"scope_type": {
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"ORGANIZATION", "ENVIRONMENT", "POPULATION"}, false),
				Required:     true,
				ForceNew:     true,
			},
			"read_only": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func resourceApplicationRoleAssignmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	envID := d.Get("environment_id").(string)
	appID := d.Get("application_id").(string)
	roleID := d.Get("role_id").(string)
	scopeID := d.Get("scope_id").(string)
	scopeType := d.Get("scope_type").(string)

	log.Printf("[INFO] Creating PingOne User Role Assignment: app %s, env %s", appID, envID)

	appRoleAssignmentRole := *pingone.NewRoleAssignmentRole(roleID)

	appRoleAssignmentScope := *pingone.NewRoleAssignmentScope(scopeID, scopeType)

	appRoleAssignment := *pingone.NewRoleAssignment(appRoleAssignmentRole, appRoleAssignmentScope) // ApplicationRoleAssignment |  (optional)

	resp, r, err := api_client.ManagementAPIsApplicationsApplicationRoleAssignmentsApi.CreateApplicationRoleAssignment(ctx, envID, appID).RoleAssignment(appRoleAssignment).Execute()
	if (err != nil) && (r.StatusCode != 201) {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsApplicationsApplicationRoleAssignmentsApi.CreateApplicationRoleAssignment``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	d.SetId(resp.GetId())

	return resourceApplicationRoleAssignmentRead(ctx, d, meta)
}

func resourceApplicationRoleAssignmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	roleAssignmentID := d.Id()
	envID := d.Get("environment_id").(string)
	appID := d.Get("application_id").(string)

	resp, r, err := api_client.ManagementAPIsApplicationsApplicationRoleAssignmentsApi.ReadOneApplicationRoleAssignment(ctx, envID, appID, roleAssignmentID).Execute()
	if err != nil {

		if r.StatusCode == 404 {
			log.Printf("[INFO] PingOne Role Assignment %s no longer exists", d.Id())
			d.SetId("")
			return nil
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsApplicationsApplicationRoleAssignmentsApi.ReadOneRoleAssignment``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	d.Set("role_id", resp.GetRole().Id)
	d.Set("scope_id", resp.GetScope().Id)
	d.Set("scope_type", resp.GetScope().Type)
	d.Set("read_only", resp.GetReadOnly())

	return diags
}

func resourceApplicationRoleAssignmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	roleAssignmentID := d.Id()
	envID := d.Get("environment_id").(string)
	appID := d.Get("application_id").(string)
	readOnly := d.Get("read_only").(bool)

	if readOnly {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Cannot be deleted, role assignment is read only"),
		})

		return diags
	}

	_, err := api_client.ManagementAPIsApplicationsApplicationRoleAssignmentsApi.DeleteApplicationRoleAssignment(ctx, envID, appID, roleAssignmentID).Execute()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsApplicationsApplicationRoleAssignmentsApi.DeleteApplicationRoleAssignment``: %v", err),
		})

		return diags
	}

	return nil
}

func resourceApplicationRoleAssignmentImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	attributes := strings.SplitN(d.Id(), "/", 3)

	if len(attributes) != 3 {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"envID/appID/roleAssignmentID\"", d.Id())
	}

	envID, appID, roleAssignmentID := attributes[0], attributes[1], attributes[2]

	d.Set("environment_id", envID)
	d.Set("application_id", appID)
	d.SetId(roleAssignmentID)

	resourceApplicationRoleAssignmentRead(ctx, d, meta)

	return []*schema.ResourceData{d}, nil
}
