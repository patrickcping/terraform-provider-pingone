package main

import (
	"context"
	"terraform-provider-pingone/pingone"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func main() {
	tfsdk.Serve(context.Background(), pingone.New, tfsdk.ServeOpts{
		Name: "pingone",
	})
}
