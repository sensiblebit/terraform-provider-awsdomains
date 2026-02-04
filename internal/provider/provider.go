package provider

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53domains"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &AWSDomainsProvider{}

type AWSDomainsProvider struct {
	version string
}

type AWSDomainsProviderModel struct {
	Region  types.String `tfsdk:"region"`
	Profile types.String `tfsdk:"profile"`
}

// ProviderData holds the AWS clients passed to resources and data sources
type ProviderData struct {
	DomainsClient *route53domains.Client
	Route53Client *route53.Client
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AWSDomainsProvider{
			version: version,
		}
	}
}

func (p *AWSDomainsProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "awsdomains"
	resp.Version = p.version
}

func (p *AWSDomainsProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
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
	}
}

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

	providerData := &ProviderData{
		DomainsClient: domainsClient,
		Route53Client: route53Client,
	}

	resp.DataSourceData = providerData
	resp.ResourceData = providerData
}

func (p *AWSDomainsProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDomainRegistrationResource,
	}
}

func (p *AWSDomainsProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDomainAvailabilityDataSource,
		NewDomainPriceDataSource,
	}
}
