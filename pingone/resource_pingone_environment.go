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

func resourceEnvironment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceEnvironmentCreate,
		ReadContext:   resourceEnvironmentRead,
		UpdateContext: resourceEnvironmentUpdate,
		DeleteContext: resourceEnvironmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceEnvironmentImport,
		},

		Schema: map[string]*schema.Schema{
			"environment_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "SANDBOX",
				ValidateFunc: validation.StringInSlice([]string{"PRODUCTION", "SANDBOX"}, false),
			},
			"region": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"NA", "EU", "AP"}, false),
				ForceNew:     true,
			},
			"license_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"product": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     billOfMaterialsProductElem,
				Set:      HashByMapKey("type"),
			},
			"default_population_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_population_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"default_population_description": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

var billOfMaterialsProductElem = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"type": {
			Type:         schema.TypeString,
			ValidateFunc: validation.StringInSlice([]string{"PING_ONE_MFA", "PING_ONE_RISK", "PING_ONE_VERIFY", "PING_ONE_BASE", "PING_FEDERATE", "PING_ACCESS", "PING_DIRECTORY", "PING_DATA_SYNC", "PING_DATA_GOVERNANCE", "PING_ONE_FOR_ENTERPRISE", "PING_ID", "PING_ID_SDK", "PING_INTELLIGENCE", "PING_CENTRAL"}, false),
			Required:     true,
		},
		"console_href": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"bookmark": {
			Type:     schema.TypeSet,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
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
	},
}

func resourceEnvironmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})

	var diags diag.Diagnostics

	envName := d.Get("name").(string)
	envDescription := d.Get("description").(string)
	envType := d.Get("type").(string)
	envRegion := d.Get("region").(string)

	log.Printf("[INFO] Creating PingOne Environment: name %s, type %s", envName, envType)

	environmentLicense := *pingone.NewEnvironmentLicense()
	if license, ok := d.GetOk("license_id"); ok {
		environmentLicense.SetId(license.(string))
	}

	environment := *pingone.NewEnvironment() // Environment |  (optional)
	environment.SetName(envName)
	environment.SetDescription(envDescription)
	environment.SetType(envType)
	environment.SetRegion(envRegion)
	environment.SetLicense(environmentLicense)

	resp, r, err := api_client.ManagementAPIsEnvironmentsApi.CreateEnvironmentActiveLicense(ctx).Environment(environment).Execute()
	if (err != nil) && (r.StatusCode != 201) {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsEnvironmentsApi.CreateEnvironmentActiveLicense``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	if products, ok := d.GetOk("product"); ok {
		productBOMItems := buildBOMProductsCreateRequest(products.(*schema.Set).List())

		billOfMaterials := *pingone.NewBillOfMaterials() // Environment |  (optional)
		billOfMaterials.SetProducts(productBOMItems)

		_, r, err := api_client.ManagementAPIsBillOfMaterialsBOMApi.UpdateBillOfMaterials(ctx, resp.GetId()).BillOfMaterials(billOfMaterials).Execute()
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Error when calling `ManagementAPIsBillOfMaterialsBOMApi.UpdateBillOfMaterials``: %v", err),
				Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
			})

			return diags
		}
	}

	//Have to create a default population because of the destroy restriction on the population resource
	popName := d.Get("default_population_name").(string)
	popDescription := d.Get("default_population_description").(string)

	log.Printf("[INFO] Creating PingOne Default Population: name %s", popName)

	population := *pingone.NewPopulation() // Population |  (optional)
	population.SetName(popName)
	population.SetDescription(popDescription)

	popResp, popR, popErr := api_client.ManagementAPIsPopulationsApi.CreatePopulation(ctx, resp.GetId()).Population(population).Execute()
	if (err != nil) && (r.StatusCode != 201) {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsPopulationsApi.CreatePopulation``: %v", popErr),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", popR.Body),
		})

		return diags
	}

	d.SetId(fmt.Sprintf("%s/%s", resp.GetId(), popResp.GetId()))

	return resourceEnvironmentRead(ctx, d, meta)
}

func resourceEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	attributes := strings.SplitN(d.Id(), "/", 2)
	envID, populationID := attributes[0], attributes[1]

	resp, r, err := api_client.ManagementAPIsEnvironmentsApi.ReadOneEnvironment(ctx, envID).Execute()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsEnvironmentsApi.ReadOneEnvironment``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	d.Set("environment_id", resp.GetId())
	d.Set("name", resp.GetName())
	d.Set("description", resp.GetDescription())
	d.Set("type", resp.GetType())
	d.Set("region", resp.GetRegion())
	d.Set("license_id", resp.GetLicense().Id)

	respBOM, rBOM, errBOM := api_client.ManagementAPIsBillOfMaterialsBOMApi.ReadOneBillOfMaterials(ctx, envID).Execute()
	if errBOM != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsEnvironmentsApi.ReadOneEnvironment``: %v", errBOM),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", rBOM),
		})

		return diags
	}

	productBOMItems := flattenBOMProducts(respBOM)
	log.Printf("products: %v\n", productBOMItems)
	d.Set("product", productBOMItems)

	popResp, popR, popErr := api_client.ManagementAPIsPopulationsApi.ReadOnePopulation(ctx, envID, populationID).Execute()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsPopulationsApi.ReadOnePopulation``: %v", popErr),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", popR.Body),
		})

		return diags
	}

	d.Set("default_population_id", popResp.GetId())
	d.Set("default_population_name", popResp.GetName())
	d.Set("default_population_description", popResp.GetDescription())

	return diags
}

func resourceEnvironmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	attributes := strings.SplitN(d.Id(), "/", 2)
	envID, populationID := attributes[0], attributes[1]
	envName := d.Get("name").(string)
	envDescription := d.Get("description").(string)
	envType := d.Get("type").(string)
	envRegion := d.Get("region").(string)

	environmentLicense := *pingone.NewEnvironmentLicense()
	if license, ok := d.GetOk("license_id"); ok {
		environmentLicense.SetId(license.(string))
	}

	environment := *pingone.NewEnvironment() // Environment |  (optional)
	environment.SetName(envName)
	environment.SetDescription(envDescription)
	environment.SetType(envType)
	environment.SetRegion(envRegion)
	environment.SetLicense(environmentLicense)

	if change := d.HasChange("type"); change {
		//If type has changed from SANDBOX -> PRODUCTION and vice versa we need a separate API call
		inlineObject2 := *pingone.NewInlineObject2()
		_, newType := d.GetChange("type")
		inlineObject2.SetType(newType.(string))
		_, r, err := api_client.ManagementAPIsEnvironmentsApi.UpdateEnvironmentType(ctx, envID).InlineObject2(inlineObject2).Execute()
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Error when calling `ManagementAPIsEnvironmentsApi.UpdateEnvironmentType``: %v", err),
				Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
			})

			return diags
		}
	}

	_, r, err := api_client.ManagementAPIsEnvironmentsApi.UpdateEnvironment(ctx, envID).Environment(environment).Execute()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsEnvironmentsApi.UpdateEnvironment``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	if products, ok := d.GetOk("product"); ok {
		productBOMItems := buildBOMProductsCreateRequest(products.(*schema.Set).List())

		billOfMaterials := *pingone.NewBillOfMaterials() // Environment |  (optional)
		billOfMaterials.SetProducts(productBOMItems)

		_, r, err := api_client.ManagementAPIsBillOfMaterialsBOMApi.UpdateBillOfMaterials(ctx, envID).BillOfMaterials(billOfMaterials).Execute()
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Error when calling `ManagementAPIsBillOfMaterialsBOMApi.UpdateBillOfMaterials``: %v", err),
				Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
			})

			return diags
		}
	}

	if change := d.HasChange("default_population_name") || d.HasChange("default_population_description"); change {

		popName := d.Get("default_population_name").(string)
		popDescription := d.Get("default_population_description").(string)

		population := *pingone.NewPopulation() // Population |  (optional)
		population.SetName(popName)
		population.SetDescription(popDescription)

		_, r, err := api_client.ManagementAPIsPopulationsApi.UpdatePopulation(ctx, envID, populationID).Population(population).Execute()
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Error when calling `ManagementAPIsPopulationsApi.UpdatePopulation``: %v", err),
				Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
			})

			return diags
		}

	}

	return resourceEnvironmentRead(ctx, d, meta)
}

func resourceEnvironmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	p1Client := meta.(*p1Client)
	api_client := p1Client.APIClient
	ctx = context.WithValue(ctx, pingone.ContextServerVariables, map[string]string{
		"suffix": p1Client.regionSuffix,
	})
	var diags diag.Diagnostics

	attributes := strings.SplitN(d.Id(), "/", 2)
	envID := attributes[0]

	_, err := api_client.ManagementAPIsEnvironmentsApi.DeleteEnvironment(ctx, envID).Execute()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsEnvironmentsApi.DeleteEnvironment``: %v", err),
		})

		return diags
	}

	return nil
}

func resourceEnvironmentImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	attributes := strings.SplitN(d.Id(), "/", 2)

	if len(attributes) != 2 {
		return nil, fmt.Errorf("invalid id (\"%s\") specified, should be in format \"envID/populationID\"", d.Id())
	}

	envID, populationID := attributes[0], attributes[1]

	d.Set("environment_id", envID)
	d.SetId(fmt.Sprintf("%s/%s", envID, populationID))

	resourceGroupRead(ctx, d, meta)

	return []*schema.ResourceData{d}, nil
}

func buildBOMProductsCreateRequest(items []interface{}) []pingone.BillOfMaterialsProducts {
	var productBOMItems []pingone.BillOfMaterialsProducts

	for _, item := range items {

		productBOM := pingone.NewBillOfMaterialsProducts()
		productBOM.SetType(item.(map[string]interface{})["type"].(string))

		log.Printf("console href %t", (item.(map[string]interface{})["console_href"] != nil) && (item.(map[string]interface{})["console_href"] != ""))

		if (item.(map[string]interface{})["console_href"] != nil) && (item.(map[string]interface{})["console_href"] != "") {
			productBOMItemConsole := pingone.NewBillOfMaterialsConsole()
			productBOMItemConsole.SetHref(item.(map[string]interface{})["console_href"].(string))

			productBOM.SetConsole(*productBOMItemConsole)
		}

		var productBOMBookmarkItems []pingone.BillOfMaterialsBookmarks

		for _, bookmarkItem := range item.(map[string]interface{})["bookmark"].(*schema.Set).List() {

			productBOMBookmark := pingone.NewBillOfMaterialsBookmarks()
			productBOMBookmark.SetName(bookmarkItem.(map[string]interface{})["name"].(string))
			productBOMBookmark.SetHref(bookmarkItem.(map[string]interface{})["href"].(string))

			productBOMBookmarkItems = append(productBOMBookmarkItems, *productBOMBookmark)
		}

		productBOM.SetBookmarks(productBOMBookmarkItems)

		productBOMItems = append(productBOMItems, *productBOM)
	}

	return productBOMItems
}

func flattenBOMProducts(items pingone.BillOfMaterials) *schema.Set {
	productItems := make([]interface{}, 0)

	if _, ok := items.GetProductsOk(); ok {

		for _, product := range items.GetProducts() {

			productItems = append(productItems, map[string]interface{}{
				"type":         product.GetType(),
				"console_href": product.Console.GetHref(),
				"bookmark":     flattenBOMProductsBookmarkList(product.GetBookmarks()),
			})

		}

	}

	return schema.NewSet(HashByMapKey("type"), productItems)
}

func flattenBOMProductsBookmarkList(bookmarkList []pingone.BillOfMaterialsBookmarks) *schema.Set {
	bookmarkItems := make([]interface{}, 0, len(bookmarkList))
	for _, bookmark := range bookmarkList {

		bookmarkName := ""
		if _, ok := bookmark.GetNameOk(); ok {
			bookmarkName = bookmark.GetName()
		}
		bookmarkHref := ""
		if _, ok := bookmark.GetHrefOk(); ok {
			bookmarkHref = bookmark.GetHref()
		}

		bookmarkItems = append(bookmarkItems, map[string]interface{}{
			"name": bookmarkName,
			"href": bookmarkHref,
		})
	}
	return schema.NewSet(HashByMapKey("name"), bookmarkItems)
}
