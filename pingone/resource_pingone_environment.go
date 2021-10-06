package pingone

//https://learn.hashicorp.com/tutorials/terraform/provider-setup?in=terraform/providers
import (
	"context"
	"fmt"
	"log"

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
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
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
			"default_population": {
				Type:     schema.TypeSet,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
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
	api_client := meta.(*pingone.APIClient)
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

	resp, r, err := api_client.ManagementAPIsEnvironmentsApi.CreateEnvironmentActiveLicense(context.Background()).Environment(environment).Execute()
	if (err != nil) && (r.StatusCode != 201) {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsEnvironmentsApi.CreateEnvironmentActiveLicense``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	d.SetId(resp.GetId())

	if products, ok := d.GetOk("product"); ok {
		productBOMItems := buildBOMProductsCreateRequest(products.(*schema.Set).List())

		billOfMaterials := *pingone.NewBillOfMaterials() // Environment |  (optional)
		billOfMaterials.SetProducts(productBOMItems)

		_, r, err := api_client.ManagementAPIsBillOfMaterialsBOMApi.UpdateBillOfMaterials(context.Background(), resp.GetId()).BillOfMaterials(billOfMaterials).Execute()
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
	popName := "Default"
	popDescription := d.Get("default_population").(*schema.Set).List()[0].(map[string]interface{})["description"].(string)

	log.Printf("[INFO] Creating PingOne Default Population: name %s", popName)

	population := *pingone.NewPopulation() // Population |  (optional)
	population.SetName(popName)
	population.SetDescription(popDescription)

	_, popR, popErr := api_client.ManagementAPIsPopulationsApi.CreatePopulation(context.Background(), resp.GetId()).Population(population).Execute()
	if (err != nil) && (r.StatusCode != 201) {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsPopulationsApi.CreatePopulation``: %v", popErr),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", popR.Body),
		})

		return diags
	}

	return resourceEnvironmentRead(ctx, d, meta)
}

func resourceEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api_client := meta.(*pingone.APIClient)
	var diags diag.Diagnostics

	envID := d.Id()

	resp, r, err := api_client.ManagementAPIsEnvironmentsApi.ReadOneEnvironment(context.Background(), envID).Execute()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsEnvironmentsApi.ReadOneEnvironment``: %v", err),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
		})

		return diags
	}

	d.Set("name", resp.GetName())
	d.Set("description", resp.GetDescription())
	d.Set("type", resp.GetType())
	d.Set("region", resp.GetRegion())
	d.Set("license_id", resp.GetLicense().Id)

	respBOM, rBOM, errBOM := api_client.ManagementAPIsBillOfMaterialsBOMApi.ReadOneBillOfMaterials(context.Background(), envID).Execute()
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

	limit := int32(1) // int32 | Adding a paging value to limit the number of resources displayed per page (optional)
	filter := "name eq \"Default\""

	popResp, popR, popErr := api_client.ManagementAPIsPopulationsApi.ReadAllPopulations(context.Background(), envID).Limit(limit).Filter(filter).Execute()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsPopulationsApi.ReadAllPopulations``: %v", popErr),
			Detail:   fmt.Sprintf("Full HTTP response: %v\n", popR.Body),
		})

		return diags
	}

	populationItems := flattenPopulations(popResp.Embedded.GetPopulations())
	if populationItems.Len() > 0 {
		log.Printf("Populations: %v\n", populationItems)
		d.Set("default_population", populationItems)
	}

	return diags
}

func resourceEnvironmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api_client := meta.(*pingone.APIClient)
	var diags diag.Diagnostics

	envID := d.Id()
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
		_, r, err := api_client.ManagementAPIsEnvironmentsApi.UpdateEnvironmentType(context.Background(), envID).InlineObject2(inlineObject2).Execute()
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Error when calling `ManagementAPIsEnvironmentsApi.UpdateEnvironmentType``: %v", err),
				Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
			})

			return diags
		}
	}

	_, r, err := api_client.ManagementAPIsEnvironmentsApi.UpdateEnvironment(context.Background(), envID).Environment(environment).Execute()
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

		_, r, err := api_client.ManagementAPIsBillOfMaterialsBOMApi.UpdateBillOfMaterials(context.Background(), envID).BillOfMaterials(billOfMaterials).Execute()
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Error when calling `ManagementAPIsBillOfMaterialsBOMApi.UpdateBillOfMaterials``: %v", err),
				Detail:   fmt.Sprintf("Full HTTP response: %v\n", r.Body),
			})

			return diags
		}
	}

	return resourceEnvironmentRead(ctx, d, meta)
}

func resourceEnvironmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api_client := meta.(*pingone.APIClient)
	var diags diag.Diagnostics

	envID := d.Id()

	_, err := api_client.ManagementAPIsEnvironmentsApi.DeleteEnvironment(context.Background(), envID).Execute()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Error when calling `ManagementAPIsEnvironmentsApi.DeleteEnvironment``: %v", err),
		})

		return diags
	}

	return nil
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

func flattenPopulations(populationList []pingone.Population) *schema.Set {
	populationItems := make([]interface{}, 0, 1)

	for _, population := range populationList {

		populationItems = append(populationItems, map[string]interface{}{
			"id":          population.GetId(),
			"name":        population.GetName(),
			"description": population.GetDescription(),
		})
	}

	return schema.NewSet(HashByMapKey("name"), populationItems)
}
