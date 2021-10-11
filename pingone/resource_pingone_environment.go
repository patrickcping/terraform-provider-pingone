package pingone

//https://learn.hashicorp.com/tutorials/terraform/provider-setup?in=terraform/providers
import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/patrickcping/pingone-go"
)

type resourceEnvironmentType struct{}

func (r resourceEnvironmentType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
			"name": {
				Type:     types.StringType,
				Required: true,
			},
			"description": {
				Type:     types.StringType,
				Optional: true,
			},
			"type": {
				Type:     types.StringType,
				Optional: true,
				// Default:      "SANDBOX",
				// ValidateFunc: validation.StringInSlice([]string{"PRODUCTION", "SANDBOX"}, false),
			},
			"region": {
				Type:     types.StringType,
				Required: true,
				// ValidateFunc: validation.StringInSlice([]string{"NA", "EU", "AP"}, false),
				// ForceNew:     true,
			},
			"license_id": {
				Type:     types.StringType,
				Required: true,
				// ForceNew: true,
			},
			"product": {
				Required: true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"type": {
						Type:     types.StringType,
						Required: true,
					},
					"console_href": {
						Type:     types.StringType,
						Optional: true,
					},
					"bookmark": {
						Optional: true,
						Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
							"name": {
								Type:     types.StringType,
								Required: true,
							},
							"href": {
								Type:     types.StringType,
								Required: true,
							},
						}, tfsdk.SetNestedAttributesOptions{}),
					},
				}, tfsdk.SetNestedAttributesOptions{}),
			},
			"default_population": {
				Required: true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Type:     types.StringType,
						Computed: true,
					},
					"name": {
						Type:     types.StringType,
						Required: true,
					},
					"description": {
						Type:     types.StringType,
						Optional: true,
					},
				}),
			},
		},
	}, nil
}

func (r resourceEnvironmentType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceEnvironment{
		p: *(p.(*provider)),
	}, nil
}

type resourceEnvironment struct {
	p provider
}

func (r resourceEnvironment) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	api_client := r.p.client

	// Retrieve values from plan
	var plan Environment
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	environmentLicense := *pingone.NewEnvironmentLicense()
	environmentLicense.SetId(plan.LicenseID.Value)

	environment := *pingone.NewEnvironment() // Environment |  (optional)
	environment.SetName(plan.Name.Value)
	environment.SetDescription(plan.Description.Value)
	if plan.Type.Null {
		environment.SetType("SANDBOX")
	} else {
		environment.SetType(plan.Type.Value)
	}
	environment.SetRegion(plan.Region.Value)
	environment.SetLicense(environmentLicense)

	log.Printf("[INFO] Creating PingOne Environment: name %s, type %s", plan.Name.Value, plan.Type.Value)

	apiResp, http, err := api_client.ManagementAPIsEnvironmentsApi.CreateEnvironmentActiveLicense(context.Background()).Environment(environment).Execute()
	if (err != nil) && (http.StatusCode != 201) {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error when calling `ManagementAPIsEnvironmentsApi.CreateEnvironmentActiveLicense``: %v", err),
			fmt.Sprintf("Full HTTP response: %v\n", http.Body),
		)

		return
	}

	productBOMItems := buildBOMProductsCreateRequest(plan.Products)

	billOfMaterials := *pingone.NewBillOfMaterials() // Environment |  (optional)
	billOfMaterials.SetProducts(productBOMItems)

	bomResp, bomHttp, bomErr := api_client.ManagementAPIsBillOfMaterialsBOMApi.UpdateBillOfMaterials(context.Background(), apiResp.GetId()).BillOfMaterials(billOfMaterials).Execute()
	if bomErr != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error when calling `ManagementAPIsBillOfMaterialsBOMApi.UpdateBillOfMaterials``: %v", bomErr),
			fmt.Sprintf("Full HTTP response: %v\n", bomHttp.Body),
		)

		return
	}

	//Have to create a default population because of the destroy restriction on the population resource
	population := *pingone.NewPopulation() // Population |  (optional)
	population.SetName(plan.DefaultPopulation.Name.Value)
	if !plan.DefaultPopulation.Description.Null {
		population.SetDescription(plan.DefaultPopulation.Description.Value)
	}

	log.Printf("[INFO] Creating PingOne Default Population: name %s", population.GetName())

	popResp, popR, popErr := api_client.ManagementAPIsPopulationsApi.CreatePopulation(context.Background(), apiResp.GetId()).Population(population).Execute()
	if (popErr != nil) || (http.StatusCode != 201) {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error when calling `ManagementAPIsPopulationsApi.CreatePopulation``: %v", popErr),
			fmt.Sprintf("Full HTTP response: %v\n", popR.Body),
		)

		return
	}

	log.Printf("Population: %s", popResp.GetName())

	// Generate resource state struct
	var result = Environment{
		ID:          types.String{Value: apiResp.GetId()},
		Name:        types.String{Value: apiResp.GetName()},
		Description: types.String{Value: apiResp.GetDescription(), Null: !apiResp.HasDescription()},
		Type:        types.String{Value: apiResp.GetType()},
		Region:      types.String{Value: apiResp.GetRegion()},
		LicenseID:   types.String{Value: *apiResp.GetLicense().Id},
		Products:    flattenBOMProducts(bomResp),
		DefaultPopulation: Population{
			ID:          types.String{Value: popResp.GetId()},
			Name:        types.String{Value: popResp.GetName()},
			Description: types.String{Value: popResp.GetDescription(), Null: !popResp.HasDescription()},
		},
	}

	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r resourceEnvironment) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	api_client := r.p.client
	log.Printf("aa")
	// Get current state
	var state Environment
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	envID := state.ID.Value
	populationID := state.DefaultPopulation.ID.Value

	apiResp, http, err := api_client.ManagementAPIsEnvironmentsApi.ReadOneEnvironment(context.Background(), envID).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error when calling `ManagementAPIsEnvironmentsApi.ReadOneEnvironment``: %v", err),
			fmt.Sprintf("Full HTTP response: %v\n", http.Body),
		)

		return
	}

	bomResp, rBOM, errBOM := api_client.ManagementAPIsBillOfMaterialsBOMApi.ReadOneBillOfMaterials(context.Background(), envID).Execute()
	if errBOM != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error when calling `ManagementAPIsEnvironmentsApi.ReadOneEnvironment``: %v", errBOM),
			fmt.Sprintf("Full HTTP response: %v\n", rBOM),
		)

		return
	}

	popResp, popR, popErr := api_client.ManagementAPIsPopulationsApi.ReadOnePopulation(context.Background(), envID, populationID).Execute()
	if popErr != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error when calling `ManagementAPIsPopulationsApi.ReadOnePopulation``: %v", popErr),
			fmt.Sprintf("Full HTTP response: %v\n", popR.Body),
		)

		return
	}

	state = Environment{
		ID:          types.String{Value: apiResp.GetId()},
		Name:        types.String{Value: apiResp.GetName()},
		Description: types.String{Value: apiResp.GetDescription(), Null: !apiResp.HasDescription()},
		Type:        types.String{Value: apiResp.GetType()},
		Region:      types.String{Value: apiResp.GetRegion()},
		LicenseID:   types.String{Value: *apiResp.GetLicense().Id},
		Products:    flattenBOMProducts(bomResp),
		DefaultPopulation: Population{
			ID:          types.String{Value: popResp.GetId()},
			Name:        types.String{Value: popResp.GetName()},
			Description: types.String{Value: popResp.GetDescription(), Null: !popResp.HasDescription()},
		},
	}
	log.Printf("a")
	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		log.Printf("b")
		return
	}
	log.Printf("c")
}

func (r resourceEnvironment) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	api_client := r.p.client

	log.Printf("z")
	// Get plan values
	var plan Environment
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("y")
	// Get current state
	var state Environment
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("x")
	envID := state.ID.Value
	populationID := state.DefaultPopulation.ID.Value

	environmentLicense := *pingone.NewEnvironmentLicense()
	environmentLicense.SetId(plan.LicenseID.Value)

	environment := *pingone.NewEnvironment() // Environment |  (optional)
	environment.SetName(plan.Name.Value)
	environment.SetDescription(plan.Description.Value)
	if plan.Type.Null {
		environment.SetType("SANDBOX")
	} else {
		environment.SetType(plan.Type.Value)
	}
	environment.SetRegion(plan.Region.Value)
	environment.SetLicense(environmentLicense)

	if change := plan.Type.Value != state.Type.Value; change {
		//If type has changed from SANDBOX -> PRODUCTION and vice versa we need a separate API call
		inlineObject2 := *pingone.NewInlineObject2()
		inlineObject2.SetType(plan.Type.Value)
		_, http, err := api_client.ManagementAPIsEnvironmentsApi.UpdateEnvironmentType(context.Background(), envID).InlineObject2(inlineObject2).Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Error when calling `ManagementAPIsEnvironmentsApi.UpdateEnvironmentType``: %v", err),
				fmt.Sprintf("Full HTTP response: %v\n", http.Body),
			)

			return
		}
	}

	apiResp, http, err := api_client.ManagementAPIsEnvironmentsApi.UpdateEnvironment(context.Background(), envID).Environment(environment).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error when calling `ManagementAPIsEnvironmentsApi.UpdateEnvironment``: %v", err),
			fmt.Sprintf("Full HTTP response: %v\n", http.Body),
		)

		return
	}

	productBOMItems := buildBOMProductsCreateRequest(plan.Products)

	billOfMaterials := *pingone.NewBillOfMaterials() // Environment |  (optional)
	billOfMaterials.SetProducts(productBOMItems)

	bomResp, bomHttp, bomErr := api_client.ManagementAPIsBillOfMaterialsBOMApi.UpdateBillOfMaterials(context.Background(), envID).BillOfMaterials(billOfMaterials).Execute()
	if bomErr != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error when calling `ManagementAPIsBillOfMaterialsBOMApi.UpdateBillOfMaterials``: %v", bomErr),
			fmt.Sprintf("Full HTTP response: %v\n", bomHttp.Body),
		)

		return
	}

	populationUpdate := *pingone.NewPopulation()
	if change := (plan.DefaultPopulation.Name.Value != state.DefaultPopulation.Name.Value) || (plan.DefaultPopulation.Description.Value != state.DefaultPopulation.Description.Value); change {

		population := *pingone.NewPopulation() // Population |  (optional)
		population.SetName(plan.DefaultPopulation.Name.Value)
		population.SetDescription(plan.DefaultPopulation.Description.Value)

		popResp, http, err := api_client.ManagementAPIsPopulationsApi.UpdatePopulation(context.Background(), envID, populationID).Population(population).Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Error when calling `ManagementAPIsPopulationsApi.UpdatePopulation``: %v", err),
				fmt.Sprintf("Full HTTP response: %v\n", http.Body),
			)

			return
		}

		log.Printf("Population updated: %s", popResp.GetId())

		populationUpdate.SetId(popResp.GetId())
		populationUpdate.SetName(popResp.GetName())
		populationUpdate.SetDescription(popResp.GetDescription())

	} else {
		populationUpdate.SetId(state.DefaultPopulation.ID.Value)
		populationUpdate.SetName(state.DefaultPopulation.Name.Value)
		populationUpdate.SetDescription(state.DefaultPopulation.Description.Value)
	}

	// Generate resource state struct
	var result = Environment{
		ID:          types.String{Value: apiResp.GetId()},
		Name:        types.String{Value: apiResp.GetName()},
		Description: types.String{Value: apiResp.GetDescription()},
		Type:        types.String{Value: apiResp.GetType()},
		Region:      types.String{Value: apiResp.GetRegion()},
		LicenseID:   types.String{Value: *apiResp.GetLicense().Id},
		Products:    flattenBOMProducts(bomResp),
		DefaultPopulation: Population{
			ID:          types.String{Value: populationUpdate.GetId()},
			Name:        types.String{Value: populationUpdate.GetName()},
			Description: types.String{Value: populationUpdate.GetDescription()},
		},
	}

	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceEnvironment) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	api_client := r.p.client

	var state Environment
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	envID := state.ID.Value

	_, err := api_client.ManagementAPIsEnvironmentsApi.DeleteEnvironment(context.Background(), envID).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error when calling `ManagementAPIsEnvironmentsApi.DeleteEnvironment``: %v", err),
			"",
		)

		return
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

func (r resourceEnvironment) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	api_client := r.p.client

	attributes := strings.SplitN(req.ID, "/", 2)

	if len(attributes) != 2 {
		resp.Diagnostics.AddError(
			fmt.Sprintf("invalid id (\"%s\") specified, should be in format \"envID/populationID\"", req.ID),
			"",
		)

		return
	}

	envID, populationID := attributes[0], attributes[1]

	apiResp, http, err := api_client.ManagementAPIsEnvironmentsApi.ReadOneEnvironment(context.Background(), envID).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error when calling `ManagementAPIsEnvironmentsApi.ReadOneEnvironment``: %v", err),
			fmt.Sprintf("Full HTTP response: %v\n", http.Body),
		)

		return
	}

	bomResp, rBOM, errBOM := api_client.ManagementAPIsBillOfMaterialsBOMApi.ReadOneBillOfMaterials(context.Background(), envID).Execute()
	if errBOM != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error when calling `ManagementAPIsEnvironmentsApi.ReadOneEnvironment``: %v", errBOM),
			fmt.Sprintf("Full HTTP response: %v\n", rBOM),
		)

		return
	}

	popResp, popR, popErr := api_client.ManagementAPIsPopulationsApi.ReadOnePopulation(context.Background(), envID, populationID).Execute()
	if popErr != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error when calling `ManagementAPIsPopulationsApi.ReadOnePopulation``: %v", popErr),
			fmt.Sprintf("Full HTTP response: %v\n", popR.Body),
		)

		return
	}

	// Generate resource state struct
	var result = Environment{
		ID:          types.String{Value: apiResp.GetId()},
		Name:        types.String{Value: apiResp.GetName()},
		Description: types.String{Value: apiResp.GetDescription()},
		Type:        types.String{Value: apiResp.GetType()},
		Region:      types.String{Value: apiResp.GetRegion()},
		LicenseID:   types.String{Value: *apiResp.GetLicense().Id},
		Products:    flattenBOMProducts(bomResp),
		DefaultPopulation: Population{
			ID:          types.String{Value: popResp.GetId()},
			Name:        types.String{Value: popResp.GetName()},
			Description: types.String{Value: popResp.GetDescription()},
		},
	}

	resp.State.Set(ctx, result)

}

func buildBOMProductsCreateRequest(items []Product) []pingone.BillOfMaterialsProducts {
	var productBOMItems []pingone.BillOfMaterialsProducts

	for _, item := range items {

		productBOM := pingone.NewBillOfMaterialsProducts()
		productBOM.SetType(item.Type.Value)

		if !item.ConsoleHref.Null {
			productBOMItemConsole := pingone.NewBillOfMaterialsConsole()
			productBOMItemConsole.SetHref(item.ConsoleHref.Value)

			productBOM.SetConsole(*productBOMItemConsole)
		}

		var productBOMBookmarkItems []pingone.BillOfMaterialsBookmarks

		for _, bookmarkItem := range item.Bookmarks {

			productBOMBookmark := pingone.NewBillOfMaterialsBookmarks()
			productBOMBookmark.SetName(bookmarkItem.Name.Value)
			productBOMBookmark.SetHref(bookmarkItem.Href.Value)

			productBOMBookmarkItems = append(productBOMBookmarkItems, *productBOMBookmark)
		}

		productBOM.SetBookmarks(productBOMBookmarkItems)

		productBOMItems = append(productBOMItems, *productBOM)
	}

	return productBOMItems
}

func flattenBOMProducts(items pingone.BillOfMaterials) []Product {
	var productItems []Product

	if _, ok := items.GetProductsOk(); ok {

		for _, product := range items.GetProducts() {

			_, consoleHrefOk := product.Console.GetHrefOk()
			productItems = append(productItems, Product{
				Type:        types.String{Value: product.GetType()},
				ConsoleHref: types.String{Value: product.Console.GetHref(), Null: !consoleHrefOk},
				Bookmarks:   flattenBOMProductsBookmarkList(product.GetBookmarks()),
			})

		}

	}

	return productItems
}

func flattenBOMProductsBookmarkList(bookmarkList []pingone.BillOfMaterialsBookmarks) []ProductBookmark {
	var bookmarkItems []ProductBookmark
	for _, bookmark := range bookmarkList {

		bookmarkName := ""
		if _, ok := bookmark.GetNameOk(); ok {
			bookmarkName = bookmark.GetName()
		}
		bookmarkHref := ""
		if _, ok := bookmark.GetHrefOk(); ok {
			bookmarkHref = bookmark.GetHref()
		}

		bookmarkItems = append(bookmarkItems, ProductBookmark{
			Name: types.String{Value: bookmarkName},
			Href: types.String{Value: bookmarkHref},
		})
	}
	return bookmarkItems
}
