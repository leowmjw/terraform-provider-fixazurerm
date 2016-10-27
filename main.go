package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/leowmjw/terraform-provider-fixazurerm/fixazurerm"
)

func main() {
	opts := plugin.ServeOpts{
		ProviderFunc: fixazurerm.Provider,
	}
	plugin.Serve(&opts)
}
