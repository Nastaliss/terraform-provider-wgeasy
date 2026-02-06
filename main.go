package main

import (
	"context"
	"log"

	"github.com/Nastaliss/terraform-provider-wgeasy/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

func main() {
	err := providerserver.Serve(context.Background(), provider.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/Nastaliss/wgeasy",
	})
	if err != nil {
		log.Fatal(err)
	}
}
