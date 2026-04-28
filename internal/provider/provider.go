// Package provider implements the Terraform provider for Route53 Domains resources.
package provider

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53domains"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &AWSDomainsProvider{}
var _ provider.ProviderWithValidateConfig = &AWSDomainsProvider{}

// AWSDomainsProvider implements the Terraform provider entrypoint.
type AWSDomainsProvider struct {
	version string
}

// AWSDomainsProviderModel stores provider configuration values.
type AWSDomainsProviderModel struct {
	Region      types.String      `tfsdk:"region"`
	Profile     types.String      `tfsdk:"profile"`
	DefaultTags *DefaultTagsModel `tfsdk:"default_tags"`
}

// DefaultTagsModel stores provider default tags configuration.
type DefaultTagsModel struct {
	Tags types.Map `tfsdk:"tags"`
}

// providerData holds the AWS clients passed to resources and data sources.
type providerData struct {
	DomainsClient *route53domains.Client
	Route53Client *route53.Client
	DefaultTags   map[string]string
}

// New returns a provider factory for Terraform.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AWSDomainsProvider{
			version: version,
		}
	}
}

// Metadata sets the provider type name and version.
func (p *AWSDomainsProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "awsdomains"
	resp.Version = p.version
}

// Schema describes the provider-level configuration.
func (p *AWSDomainsProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provider for managing AWS Route53 domain registrations with full lifecycle support.",
		Attributes: map[string]schema.Attribute{
			"region": schema.StringAttribute{
				Description: "AWS region for Route53 Domains API (must be us-east-1).",
				Optional:    true,
			},
			"profile": schema.StringAttribute{
				Description: "AWS profile to use for authentication.",
				Optional:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"default_tags": schema.SingleNestedBlock{
				Description: "Default tags to apply to all taggable resources managed by this provider.",
				Attributes: map[string]schema.Attribute{
					"tags": schema.MapAttribute{
						Optional:    true,
						ElementType: types.StringType,
						Description: "Map of default tag keys and values. Resource-level tags override default tags with the same key.",
					},
				},
			},
		},
	}
}

// ValidateConfig validates provider-level configuration before AWS clients are configured.
func (p *AWSDomainsProvider) ValidateConfig(ctx context.Context, req provider.ValidateConfigRequest, resp *provider.ValidateConfigResponse) {
	var data AWSDomainsProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.DefaultTags == nil || !frameworkMapElementsKnown(data.DefaultTags.Tags) {
		return
	}

	defaultTags, diags := frameworkMapToStringMap(ctx, data.DefaultTags.Tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	addTagValidationDiagnostics(&resp.Diagnostics, path.Root("default_tags").AtName("tags"), defaultTags)
}

// Configure creates AWS service clients and shares them with resources and data sources.
func (p *AWSDomainsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data AWSDomainsProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build AWS config options
	var optFns []func(*config.LoadOptions) error

	// Route53 Domains API only works in us-east-1
	region := "us-east-1"
	if !data.Region.IsNull() {
		region = data.Region.ValueString()
	}
	optFns = append(optFns, config.WithRegion(region))

	if !data.Profile.IsNull() {
		optFns = append(optFns, config.WithSharedConfigProfile(data.Profile.ValueString()))
	}

	defaultTags := map[string]string{}
	if data.DefaultTags != nil {
		convertedDefaultTags, diags := frameworkMapToStringMap(ctx, data.DefaultTags.Tags)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		addTagValidationDiagnostics(&resp.Diagnostics, path.Root("default_tags").AtName("tags"), convertedDefaultTags)
		if resp.Diagnostics.HasError() {
			return
		}
		defaultTags = convertedDefaultTags
	}

	cfg, err := config.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create AWS config",
			"An error occurred while creating the AWS configuration: "+err.Error(),
		)
		return
	}

	domainsClient := route53domains.NewFromConfig(cfg)
	route53Client := route53.NewFromConfig(cfg)

	providerData := &providerData{
		DomainsClient: domainsClient,
		Route53Client: route53Client,
		DefaultTags:   defaultTags,
	}

	resp.DataSourceData = providerData
	resp.ResourceData = providerData
}

// Resources returns the provider resources.
func (p *AWSDomainsProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDomainRegistrationResource,
	}
}

// DataSources returns the provider data sources.
func (p *AWSDomainsProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDomainAvailabilityDataSource,
		NewDomainPriceDataSource,
	}
}
