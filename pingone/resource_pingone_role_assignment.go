package pingone

//https://learn.hashicorp.com/tutorials/terraform/provider-setup?in=terraform/providers
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

func resourceUserRoleAssignment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserRoleAssignmentCreate,
		ReadContext:   resourceUserRoleAssignmentRead,
		DeleteContext: resourceUserRoleAssignmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceUserRoleAssignmentImport,
		},

		Schema: map[string]*schema.Schema{
			"environment_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"user_id": {
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

func resourceUserRoleAssignmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerIndex, p1Client.regionUrlIndex)
	var diags diag.Diagnostics

	envID := d.Get("environment_id").(string)
	userID := d.Get("user_id").(string)
	roleID := d.Get("role_id").(string)
	scopeID := d.Get("scope_id").(string)
	scopeType := d.Get("scope_type").(string)

	log.Printf("[INFO] Creating PingOne User Role Assignment: user %s, env %s", userID, envID)

	userRoleAssignmentRole := *pingone.NewRoleAssignmentRole()
	userRoleAssignmentRole.SetId(roleID)

	userRoleAssignmentScope := *pingone.NewRoleAssignmentScope()
	userRoleAssignmentScope.SetId(scopeID)
	userRoleAssignmentScope.SetType(scopeType)

	userRoleAssignment := *pingone.NewRoleAssignment() // UserRoleAssignment |  (optional)
	userRoleAssignment.SetRole(userRoleAssignmentRole)
	userRoleAssignment.SetScope(userRoleAssignmentScope)

	resp, r, err := api_client.ManagementAPIsUsersUserRoleAssignmentsApi.CreateUserRoleAssignment(ctx, envID, userID).RoleAssignment(userRoleAssignment).Execute()
	if (err != nil) && (r.StatusCode != 201) {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsUsersUserRoleAssignmentsApi.CreateUserRoleAssignment``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	d.SetId(resp.GetId())

	return resourceUserRoleAssignmentRead(ctx, d, meta)
}

func resourceUserRoleAssignmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerIndex, p1Client.regionUrlIndex)
	var diags diag.Diagnostics

	roleAssignmentID := d.Id()
	envID := d.Get("environment_id").(string)
	userID := d.Get("user_id").(string)

	resp, r, err := api_client.ManagementAPIsUsersUserRoleAssignmentsApi.ReadOneRoleAssignment(ctx, envID, userID, roleAssignmentID).Execute()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsUsersUserRoleAssignmentsApi.ReadOneRoleAssignment``: %v", err),
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

func resourceUserRoleAssignmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerIndex, p1Client.regionUrlIndex)
	var diags diag.Diagnostics

	roleAssignmentID := d.Id()
	envID := d.Get("environment_id").(string)
	userID := d.Get("user_id").(string)
	readOnly := d.Get("read_only").(bool)

	if readOnly {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Cannot be deleted, role assignment is read only"),
		})

		return diags
	}

	_, err := api_client.ManagementAPIsUsersUserRoleAssignmentsApi.DeleteUserRoleAssignment(ctx, envID, userID, roleAssignmentID).Execute()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsUsersUserRoleAssignmentsApi.DeleteUserRoleAssignment``: %v", err),
		})

		return diags
	}

	return nil
}

func resourceUserRoleAssignmentImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	attributes := strings.SplitN(d.Id(), "/", 3)

	if len(attributes) != 3 {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"envID/userID/roleAssignmentID\"", d.Id())
	}

	envID, userID, roleAssignmentID := attributes[0], attributes[1], attributes[2]

	d.Set("environment_id", envID)
	d.Set("user_id", userID)
	d.SetId(roleAssignmentID)

	resourceUserRoleAssignmentRead(ctx, d, meta)

	return []*schema.ResourceData{d}, nil
}
