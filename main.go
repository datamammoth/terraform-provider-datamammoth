package main

import (
	"context"
	"github.com/datamammoth/terraform-provider-datamammoth/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

func main() {
	providerserver.Serve(context.Background(), provider.New("0.1.0"), providerserver.ServeOpts{
		Address: "registry.terraform.io/datamammoth/datamammoth",
	})
}
