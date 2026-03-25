// Root-level entry point for terraform-plugin-docs generation.
// The canonical build entry point is cmd/terraform-provider-pdnd/main.go.
package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/lmammino/terraform-provider-pdnd/internal/provider"
)

var version = "dev"

func main() {
	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/lmammino/pdnd",
	}
	err := providerserver.Serve(context.Background(), provider.New(version), opts)
	if err != nil {
		log.Fatal(err)
	}
}
