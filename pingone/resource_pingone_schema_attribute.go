package pingone

//https://learn.hashicorp.com/tutorials/terraform/provider-setup?in=terraform/providers
import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/patrickcping/pingone-go"
)

func resourceSchemaAttribute() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSchemaAttributeCreate,
		ReadContext:   resourceSchemaAttributeRead,
		UpdateContext: resourceSchemaAttributeUpdate,
		DeleteContext: resourceSchemaAttributeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceSchemaAttributeImport,
		},

		Schema: map[string]*schema.Schema{
			"environment_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"schema_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"display_name": {
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
				Default:  true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "STRING",
			},
			"unique": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"multivalued": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"ldap_attribute": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"required": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"schema_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceSchemaAttributeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerIndex, p1Client.regionUrlIndex)
	var diags diag.Diagnostics

	envID := d.Get("environment_id").(string)
	schemaID := d.Get("schema_id").(string)
	attributeName := d.Get("name").(string)
	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	enabled := d.Get("enabled").(bool)
	attributeType := d.Get("type").(string)
	unique := d.Get("unique").(bool)
	multivalued := d.Get("multivalued").(bool)
	required := d.Get("required").(bool)

	log.Printf("[INFO] Creating PingOne Schema Attribute: name %s", attributeName)

	schemaAttribute := *pingone.NewSchemaAttribute() // SchemaAttribute |  (optional)
	schemaAttribute.SetName(attributeName)
	schemaAttribute.SetDisplayName(displayName)
	schemaAttribute.SetDescription(description)
	schemaAttribute.SetEnabled(enabled)
	schemaAttribute.SetType(attributeType)
	schemaAttribute.SetUnique(unique)
	schemaAttribute.SetMultiValued(multivalued)
	schemaAttribute.SetRequired(required)

	resp, r, err := api_client.ManagementAPIsSchemasApi.CreateAttribute(ctx, envID, schemaID).SchemaAttribute(schemaAttribute).Execute()
	if (err != nil) && (r.StatusCode != 201) {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsSchemasApi.CreateAttribute``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	d.SetId(resp.GetId())

	return resourceSchemaAttributeRead(ctx, d, meta)
}

func resourceSchemaAttributeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerIndex, p1Client.regionUrlIndex)
	var diags diag.Diagnostics

	schemaAttributeID := d.Id()
	envID := d.Get("environment_id").(string)
	schemaID := d.Get("schema_id").(string)

	resp, r, err := api_client.ManagementAPIsSchemasApi.ReadOneAttribute(ctx, envID, schemaID, schemaAttributeID).Execute()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsSchemasApi.ReadOneAttribute``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	d.Set("name", resp.GetName())
	d.Set("display_name", resp.GetDisplayName())
	d.Set("description", resp.GetDescription())
	d.Set("enabled", resp.GetEnabled())
	d.Set("type", resp.GetType())
	d.Set("unique", resp.GetUnique())
	d.Set("multivalued", resp.GetMultiValued())
	d.Set("ldap_attribute", resp.GetLdapAttribute())
	d.Set("required", resp.GetRequired())
	d.Set("schema_type", resp.GetSchemaType())

	return diags
}

func resourceSchemaAttributeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerIndex, p1Client.regionUrlIndex)
	var diags diag.Diagnostics

	attributeID := d.Id()

	envID := d.Get("environment_id").(string)
	schemaID := d.Get("schema_id").(string)
	attributeName := d.Get("name").(string)
	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	enabled := d.Get("enabled").(bool)
	attributeType := d.Get("type").(string)
	unique := d.Get("unique").(bool)
	multivalued := d.Get("multivalued").(bool)
	required := d.Get("required").(bool)

	log.Printf("[INFO] Updating PingOne Schema Attribute: name %s", attributeName)

	schemaAttribute := *pingone.NewSchemaAttribute() // SchemaAttribute |  (optional)
	schemaAttribute.SetName(attributeName)
	schemaAttribute.SetDisplayName(displayName)
	schemaAttribute.SetDescription(description)
	schemaAttribute.SetEnabled(enabled)
	schemaAttribute.SetType(attributeType)
	schemaAttribute.SetUnique(unique)
	schemaAttribute.SetMultiValued(multivalued)
	schemaAttribute.SetRequired(required)

	_, r, err := api_client.ManagementAPIsSchemasApi.UpdateAttributePatch(ctx, envID, schemaID, attributeID).SchemaAttribute(schemaAttribute).Execute()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsSchemasApi.UpdateAttributePatch``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	return resourceSchemaAttributeRead(ctx, d, meta)
}

func resourceSchemaAttributeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerIndex, p1Client.regionUrlIndex)
	var diags diag.Diagnostics

	envID := d.Get("environment_id").(string)
	schemaID := d.Get("schema_id").(string)

	attributeID := d.Id()

	_, err := api_client.ManagementAPIsSchemasApi.DeleteAttribute(ctx, envID, schemaID, attributeID).Execute()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsSchemasApi.DeleteAttribute``: %v", err),
		})

		return diags
	}

	return nil
}

func resourceSchemaAttributeImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	attributes := strings.SplitN(d.Id(), "/", 3)

	if len(attributes) != 2 {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"envID/schemaID/attributeID\"", d.Id())
	}

	envID, schemaID, attributeID := attributes[0], attributes[1], attributes[2]

	d.Set("environment_id", envID)
	d.Set("schema_id", schemaID)
	d.SetId(attributeID)

	resourceSchemaAttributeRead(ctx, d, meta)

	return []*schema.ResourceData{d}, nil
}
