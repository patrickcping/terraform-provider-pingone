package pingone

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/patrickcping/pingone-go"
)

func resourceApplicationOIDC() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceApplicationOIDCCreate,
		ReadContext:   resourceApplicationOIDCRead,
		UpdateContext: resourceApplicationOIDCUpdate,
		DeleteContext: resourceApplicationOIDCDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceApplicationOIDCImport,
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
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"secret": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"protocol": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"PING_FED_CONNECTION_INTEGRATION"}, false),
				},
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"WEB_APP", "NATIVE_APP", "SINGLE_PAGE_APP", "WORKER"}, false),
			},
			"home_page_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"login_page_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"support_unsigned_request_object": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"grant_types": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"AUTHORIZATION_CODE", "IMPLICIT", "REFRESH_TOKEN", "CLIENT_CREDENTIALS"}, false),
				},
				Required: true,
			},
			"response_types": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"TOKEN", "ID_TOKEN", "CODE"}, false),
				},
				Optional: true,
			},
			"token_endpoint_authn_method": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"NONE", "CLIENT_SECRET_BASIC", "CLIENT_SECRET_POST"}, false),
			},
			"pkce_enforcement": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "OPTIONAL",
				ValidateFunc: validation.StringInSlice([]string{"OPTIONAL", "REQUIRED", "S256_REQUIRED"}, false),
			},
			"redirect_uris": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
			"post_logout_redirect_uris": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
			"refresh_token_duration": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  2592000,
			},
			"refresh_token_rolling_duration": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  15552000,
			},
			"assign_actor_roles": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"access_control": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"role_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"ADMIN_USERS_ONLY"}, false),
						},
						"group": {
							Type:     schema.TypeSet,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"ANY_GROUP", "ALL_GROUPS"}, false),
									},
									"groups": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"icon": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"href": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"mobile": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bundle_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"package_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"integrity_detection": {
							Type:     schema.TypeSet,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"mode": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"cache_duration": {
										Type:     schema.TypeSet,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"amount": {
													Type:     schema.TypeInt,
													Required: true,
												},
												"units": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice([]string{"MINUTES", "HOURS"}, false),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"bundle_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"package_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceApplicationOIDCCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	envID := d.Get("environment_id").(string)

	application, err := expandApplicationOIDC(d)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Cannot expand Application into SDK object",
			Detail:   fmt.Sprintf("Full error: %v\n", err),
		})

		return diags
	}

	log.Printf("[INFO] Creating PingOne Application: name %s", application.GetName())

	resp, r, err := api_client.ManagementAPIsApplicationsApplicationsApi.CreateApplication(ctx, envID).OneOfApplicationSAMLApplicationOIDC(application).Execute()
	if (err != nil) || (r.StatusCode != 201) {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsApplicationsApplicationsApi.CreateApplication``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	appID := resp.(map[string]interface{})["id"].(string)

	// The platform pre-assigns roles on creation.  We should clear these down as these should be explicitly managed by TF

	// We'll be racing the platform here, so we'll loop for 10 seconds or until we find no more created

	var i int
	for start := time.Now(); time.Since(start) < (10 * time.Second); {

		log.Printf("[DEBUG] Role assignments iteration: %v", i)

		respAR, r, err := api_client.ManagementAPIsApplicationsApplicationRoleAssignmentsApi.ReadApplicationRoleAssignments(ctx, envID, appID).Execute()
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("Error when calling `ManagementAPIsApplicationsApplicationRoleAssignmentsApi.ReadApplicationRoleAssignments``: %v", err),
				Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
			})
		}

		if _, ok := respAR.Embedded.GetRoleAssignmentsOk(); ok {

			roleAssignments := respAR.Embedded.GetRoleAssignments()

			log.Printf("[DEBUG] Role assignments to process: %v", len(roleAssignments))

			// Break the loop as it appears no more role assignments got created by the platform
			if len(roleAssignments) == 0 {
				break
			}

			for _, roleAssignment := range roleAssignments {

				if !roleAssignment.GetReadOnly() {

					_, err := api_client.ManagementAPIsApplicationsApplicationRoleAssignmentsApi.DeleteApplicationRoleAssignment(ctx, envID, appID, roleAssignment.GetId()).Execute()
					if err != nil {
						log.Printf("Error %v", err)
						diags = append(diags, diag.Diagnostic{
							Severity: diag.Warning,
							Summary:  fmt.Sprintf("Error when calling `ManagementAPIsApplicationsApplicationRoleAssignmentsApi.DeleteApplicationRoleAssignment``: %v", err),
						})
					}

				}

			}

		}

		i++

		// We'll also use a max iteration of 11 loops as backstop to avoid endless looping
		if i > 10 {
			log.Println("[WARN] Loop to remove role assignments hit it's backstop value")
			break
		}
	}

	d.SetId(appID)

	return resourceApplicationOIDCRead(ctx, d, meta)
}

func resourceApplicationOIDCRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	appID := d.Id()
	envID := d.Get("environment_id").(string)

	resp, r, err := api_client.ManagementAPIsApplicationsApplicationsApi.ReadOneApplication(ctx, envID, appID).Execute()
	if err != nil {

		if r.StatusCode == 404 {
			log.Printf("[INFO] PingOne Application no %s longer exists", d.Id())
			d.SetId("")
			return nil
		}

		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsApplicationsApplicationsApi.ReadOneApplication``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	respSecret, r, err := api_client.ManagementAPIsApplicationsApplicationSecretApi.ReadApplicationSecret(ctx, envID, appID).Execute()
	if err != nil {

		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsApplicationsApplicationSecretApi.ReadApplicationSecret``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	b, err := json.Marshal(resp)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Cannot marshal application json to byte",
			Detail:   fmt.Sprintf("Full response: %v\n", err),
		})

		return diags
	}

	application := pingone.ApplicationOIDC{}
	json.Unmarshal([]byte(b), &application)

	d.Set("name", application.GetName())
	d.Set("description", application.GetDescription())
	d.Set("enabled", application.GetEnabled())
	d.Set("protocol", application.GetProtocol())
	d.Set("secret", respSecret.GetSecret())
	d.Set("tags", application.GetTags())
	d.Set("type", application.GetType())
	d.Set("home_page_url", application.GetHomePageUrl())
	d.Set("login_page_url", application.GetLoginPageUrl())
	d.Set("support_unsigned_request_object", application.GetSupportUnsignedRequestObject())
	d.Set("grant_types", application.GetGrantTypes())
	d.Set("response_types", application.GetResponseTypes())
	d.Set("token_endpoint_authn_method", application.GetTokenEndpointAuthMethod())
	d.Set("pkce_enforcement", application.GetPkceEnforcement())
	d.Set("redirect_uris", application.GetRedirectUris())
	d.Set("post_logout_redirect_uris", application.GetPostLogoutRedirectUris())
	d.Set("refresh_token_duration", application.GetRefreshTokenDuration())
	d.Set("refresh_token_rolling_duration", application.GetRefreshTokenRollingDuration())
	d.Set("assign_actor_roles", application.GetAssignActorRoles())
	d.Set("bundle_id", application.GetBundleId())
	d.Set("package_name", application.GetPackageName())

	if v, ok := application.GetAccessControlOk(); ok {

		accessControlFlattened, err := flattenApplicationAccessControl(v)

		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Cannot flatten Application from SDK object",
				Detail:   fmt.Sprintf("Full error: %v\n", err),
			})

			return diags
		}
		d.Set("access_control", accessControlFlattened)
	}

	if v, ok := application.GetIconOk(); ok {
		iconFlattened, err := flattenApplicationIcon(v)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Cannot flatten Application from SDK object",
				Detail:   fmt.Sprintf("Full error: %v\n", err),
			})

			return diags
		}
		d.Set("icon", iconFlattened)
	}

	if v, ok := application.GetMobileOk(); ok {
		mobileFlattened, err := flattenApplicationMobile(v)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Cannot flatten Application from SDK object",
				Detail:   fmt.Sprintf("Full error: %v\n", err),
			})

			return diags
		}
		d.Set("mobile", mobileFlattened)
	}

	return diags
}

func resourceApplicationOIDCUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	appID := d.Id()

	envID := d.Get("environment_id").(string)

	application, err := expandApplicationOIDC(d)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Cannot expand Application into SDK object",
			Detail:   fmt.Sprintf("Full error: %v\n", err),
		})

		return diags
	}

	log.Printf("[INFO] Updating PingOne Application: name %s", application.GetName())

	_, r, err := api_client.ManagementAPIsApplicationsApplicationsApi.UpdateApplication(ctx, envID, appID).OneOfApplicationSAMLApplicationOIDC(application).Execute()
	if err != nil {

		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsApplicationsApplicationsApi.UpdateApplication``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	return resourceApplicationOIDCRead(ctx, d, meta)
}

func resourceApplicationOIDCDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	envID := d.Get("environment_id").(string)

	appID := d.Id()

	_, err := api_client.ManagementAPIsApplicationsApplicationsApi.DeleteApplication(ctx, envID, appID).Execute()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsApplicationsApplicationsApi.DeleteApplication``: %v", err),
		})

		return diags
	}

	return nil
}

func resourceApplicationOIDCImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	attributes := strings.SplitN(d.Id(), "/", 3)

	if len(attributes) != 2 {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"envID/appID\"", d.Id())
	}

	envID, appID := attributes[0], attributes[1]

	d.Set("environment_id", envID)
	d.SetId(appID)

	resourceApplicationOIDCRead(ctx, d, meta)

	return []*schema.ResourceData{d}, nil
}

func flattenApplicationAccessControl(in *pingone.ApplicationAccessControl) ([]interface{}, error) {

	var flattenedApplicationAccessControlRole []interface{}
	var flattenedApplicationAccessControlGroup []interface{}

	if v, ok := in.GetRoleOk(); ok {
		if v1, ok := v.GetTypeOk(); ok {
			flattenedApplicationAccessControlRole = append(flattenedApplicationAccessControlRole, map[string]interface{}{
				"role_type": v1,
			})
		}
	}

	if v, ok := in.GetGroupOk(); ok {

		var flattenedApplicationAccessControlGroupMap []interface{}
		if v1, ok := v.GetTypeOk(); ok {

			groupItems := make([]interface{}, 0, len(v.GetGroups()))
			for _, group := range v.GetGroups() {

				groupItems = append(groupItems, group.GetId())
			}

			flattenedApplicationAccessControlGroupMap = append(flattenedApplicationAccessControlGroupMap, map[string]interface{}{
				"type":   v1,
				"groups": groupItems,
			})

		}

		flattenedApplicationAccessControlGroup = append(flattenedApplicationAccessControlGroup, map[string]interface{}{
			"group": flattenedApplicationAccessControlGroupMap,
		})

	}

	items := make([]interface{}, 0)
	items = append(items, map[string]interface{}{
		"role":  flattenedApplicationAccessControlRole,
		"group": flattenedApplicationAccessControlGroup,
	})

	return items, nil
}

func flattenApplicationIcon(in *pingone.ApplicationIcon) ([]interface{}, error) {

	items := make([]interface{}, 0)
	items = append(items, map[string]interface{}{
		"id":   in.GetId(),
		"href": in.GetHref(),
	})

	return items, nil
}

func flattenApplicationMobile(in *pingone.ApplicationOIDCAllOfMobile) ([]interface{}, error) {

	var flattenApplicationMobileIntegrityDetection []interface{}

	items := make([]interface{}, 0)
	items = append(items, map[string]interface{}{
		"bundle_id":           in.GetBundleId(),
		"package_name":        in.GetPackageName(),
		"integrity_detection": flattenApplicationMobileIntegrityDetection,
	})

	return items, nil
}

func expandApplicationOIDC(d *schema.ResourceData) (pingone.ApplicationOIDC, error) {

	grantTypes := marshalInterfaceToString(d.Get("grant_types").([]interface{}))

	application := *pingone.NewApplicationOIDC(d.Get("enabled").(bool), d.Get("name").(string), "OPENID_CONNECT", d.Get("type").(string), grantTypes, d.Get("token_endpoint_authn_method").(string)) // ApplicationOIDC |  (optional)
	if v, ok := d.GetOk("description"); ok {
		application.SetDescription(v.(string))
	}

	if v, ok := d.GetOk("tags"); ok {
		application.SetTags(marshalInterfaceToString(v.([]interface{})))
	}

	if v, ok := d.GetOk("home_page_url"); ok {
		application.SetHomePageUrl(v.(string))
	}

	if v, ok := d.GetOk("login_page_url"); ok {
		application.SetLoginPageUrl(v.(string))
	}

	if v, ok := d.GetOk("support_unsigned_request_object"); ok {
		application.SetSupportUnsignedRequestObject(v.(bool))
	}

	if v, ok := d.GetOk("response_types"); ok {
		application.SetResponseTypes(marshalInterfaceToString(v.([]interface{})))
	}

	if v, ok := d.GetOk("pkce_enforcement"); ok {
		application.SetPkceEnforcement(v.(string))
	}

	if v, ok := d.GetOk("redirect_uris"); ok {
		application.SetRedirectUris(marshalInterfaceToString(v.([]interface{})))
	}

	if v, ok := d.GetOk("post_logout_redirect_uris"); ok {
		application.SetPostLogoutRedirectUris(marshalInterfaceToString(v.([]interface{})))
	}

	if v, ok := d.GetOk("refresh_token_duration"); ok {
		application.SetRefreshTokenDuration(int32(v.(int)))
	}

	if v, ok := d.GetOk("refresh_token_rolling_duration"); ok {
		application.SetRefreshTokenRollingDuration(int32(v.(int)))
	}

	if v, ok := d.GetOk("assign_actor_roles"); ok {
		application.SetAssignActorRoles(v.(bool))
	}

	accessControl := *pingone.NewApplicationAccessControl()

	if v, ok := d.GetOk("access_control"); ok {

		accessControlGroupIn := v.(*schema.Set).List()[0].(map[string]interface{})["group"]

		if accessControlGroupIn != nil && len(accessControlGroupIn.(*schema.Set).List()) > 0 {

			groupsIn := accessControlGroupIn.(*schema.Set).List()[0].(map[string]interface{})["groups"].([]interface{})
			groupItems := make([]pingone.ApplicationAccessControlGroupGroups, 0, len(groupsIn))
			for _, group := range groupsIn {
				groupItems = append(groupItems, pingone.ApplicationAccessControlGroupGroups{
					Id: group.(string),
				})
			}

			accessControlGroup := *pingone.NewApplicationAccessControlGroup(
				accessControlGroupIn.(*schema.Set).List()[0].(map[string]interface{})["type"].(string),
				groupItems,
			)
			accessControl.SetGroup(accessControlGroup)

		}

		accessControlRoleIn := v.(*schema.Set).List()[0].(map[string]interface{})["role_type"]

		if accessControlRoleIn != nil && accessControlRoleIn != "" {

			accessControlRole := *pingone.NewApplicationAccessControlRole(
				accessControlRoleIn.(string),
			)
			accessControl.SetRole(accessControlRole)

		}

	}
	application.SetAccessControl(accessControl)

	if v, ok := d.GetOk("icon"); ok {

		icon := *pingone.NewApplicationIcon(v.([]interface{})[0].(map[string]interface{})["id"].(string), v.([]interface{})[0].(map[string]interface{})["href"].(string))

		application.SetIcon(icon)
	}

	// if v, ok := d.GetOk("mobile"); ok {

	// 	mobileIntegrityDetectionCacheDuration := *pingone.NewApplicationOIDCAllOfMobileIntegrityDetectionCacheDuration()
	// 	mobileIntegrityDetectionCacheDuration.SetAmount(v.([]map[string]interface{})[0]["integrity_detection"].([]map[string]interface{})[0]["cache_duration"].([]map[string]interface{})[0]["amount"].(int32))
	// 	mobileIntegrityDetectionCacheDuration.SetUnits(v.([]map[string]interface{})[0]["integrity_detection"].([]map[string]interface{})[0]["cache_duration"].([]map[string]interface{})[0]["units"].(string))

	// 	mobileIntegrityDetection := *pingone.NewApplicationOIDCAllOfMobileIntegrityDetection()
	// 	mobileIntegrityDetection.SetMode(v.([]map[string]interface{})[0]["integrity_detection"].([]map[string]interface{})[0]["mode"].(string))
	// 	mobileIntegrityDetection.SetCacheDuration(mobileIntegrityDetectionCacheDuration)

	// 	mobile := *pingone.NewApplicationOIDCAllOfMobile()
	// 	mobile.SetBundleId(v.([]map[string]interface{})[0]["bundle_id"].(string))
	// 	mobile.SetPackageName(v.([]map[string]interface{})[0]["package_name"].(string))
	// 	mobile.SetIntegrityDetection(mobileIntegrityDetection)

	// 	application.SetMobile(mobile)
	// }

	if v, ok := d.GetOk("bundle_id"); ok {
		application.SetBundleId(v.(string))
	}

	if v, ok := d.GetOk("package_name"); ok {
		application.SetPackageName(v.(string))
	}

	return application, nil

}
