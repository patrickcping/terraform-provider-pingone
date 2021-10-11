package pingone

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Environment -
type Environment struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	Type              types.String `tfsdk:"type"`
	Region            types.String `tfsdk:"region"`
	LicenseID         types.String `tfsdk:"license_id"`
	Products          []Product    `tfsdk:"product"`
	DefaultPopulation Population   `tfsdk:"default_population"`
}

type Product struct {
	Type        types.String      `tfsdk:"type"`
	ConsoleHref types.String      `tfsdk:"console_href"`
	Bookmarks   []ProductBookmark `tfsdk:"bookmark"`
}

type ProductBookmark struct {
	Name types.String `tfsdk:"name"`
	Href types.String `tfsdk:"href"`
}

type Population struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}
