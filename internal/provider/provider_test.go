package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	providerschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
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
	defaultTagsBlock, ok := resp.Schema.Blocks["default_tags"]
	if !ok {
		t.Error("Schema missing 'default_tags' block")
	}
	defaultTagsNestedBlock, ok := defaultTagsBlock.(providerschema.SingleNestedBlock)
	if !ok {
		t.Fatalf("Schema 'default_tags' block has type %T, want schema.SingleNestedBlock", defaultTagsBlock)
	}
	defaultTagsTagsAttr, ok := defaultTagsNestedBlock.Attributes["tags"]
	if !ok {
		t.Fatal("Schema 'default_tags' block missing 'tags' attribute")
	}
	defaultTagsTagsMapAttr, ok := defaultTagsTagsAttr.(providerschema.MapAttribute)
	if !ok {
		t.Fatalf("Schema 'default_tags.tags' attribute has type %T, want schema.MapAttribute", defaultTagsTagsAttr)
	}
	if !defaultTagsTagsMapAttr.Optional {
		t.Error("Schema 'default_tags.tags' attribute must be optional")
	}
	if defaultTagsTagsMapAttr.Required {
		t.Error("Schema 'default_tags.tags' attribute must not be required")
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
