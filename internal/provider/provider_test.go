package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"awsdomains": providerserver.NewProtocol6WithError(New("test")()),
}

func TestProviderSchema(t *testing.T) {
	ctx := context.Background()
	p := New("test")()

	req := provider.SchemaRequest{}
	resp := &provider.SchemaResponse{}
	p.Schema(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema returned errors: %v", resp.Diagnostics)
	}

	// Verify expected attributes exist
	attrs := resp.Schema.Attributes
	if _, ok := attrs["region"]; !ok {
		t.Error("Schema missing 'region' attribute")
	}
	if _, ok := attrs["profile"]; !ok {
		t.Error("Schema missing 'profile' attribute")
	}
}

func TestProviderMetadata(t *testing.T) {
	ctx := context.Background()
	p := New("1.0.0")()

	req := provider.MetadataRequest{}
	resp := &provider.MetadataResponse{}
	p.Metadata(ctx, req, resp)

	if resp.TypeName != "awsdomains" {
		t.Errorf("Expected TypeName 'awsdomains', got '%s'", resp.TypeName)
	}
	if resp.Version != "1.0.0" {
		t.Errorf("Expected Version '1.0.0', got '%s'", resp.Version)
	}
}
