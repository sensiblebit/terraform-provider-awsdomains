package provider

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53domains"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &DomainAvailabilityDataSource{}

type DomainAvailabilityDataSource struct {
	client *route53domains.Client
}

type DomainAvailabilityDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	DomainName   types.String `tfsdk:"domain_name"`
	Availability types.String `tfsdk:"availability"`
	Available    types.Bool   `tfsdk:"available"`
}

func NewDomainAvailabilityDataSource() datasource.DataSource {
	return &DomainAvailabilityDataSource{}
}

func (d *DomainAvailabilityDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_availability"
}

func (d *DomainAvailabilityDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Check if a domain name is available for registration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The domain name.",
			},
			"domain_name": schema.StringAttribute{
				Required:    true,
				Description: "The domain name to check availability for.",
			},
			"availability": schema.StringAttribute{
				Computed:    true,
				Description: "The availability status: AVAILABLE, AVAILABLE_RESERVED, AVAILABLE_PREORDER, UNAVAILABLE, UNAVAILABLE_PREMIUM, UNAVAILABLE_RESTRICTED, RESERVED, DONT_KNOW.",
			},
			"available": schema.BoolAttribute{
				Computed:    true,
				Description: "True if the domain is available for registration.",
			},
		},
	}
}

func (d *DomainAvailabilityDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*route53domains.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *route53domains.Client, got: %T", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *DomainAvailabilityDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DomainAvailabilityDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainName := data.DomainName.ValueString()

	output, err := d.client.CheckDomainAvailability(ctx, &route53domains.CheckDomainAvailabilityInput{
		DomainName: aws.String(domainName),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error checking domain availability",
			fmt.Sprintf("Could not check availability for %s: %s", domainName, err.Error()),
		)
		return
	}

	data.ID = types.StringValue(domainName)
	data.Availability = types.StringValue(string(output.Availability))
	data.Available = types.BoolValue(output.Availability == "AVAILABLE" || output.Availability == "AVAILABLE_RESERVED" || output.Availability == "AVAILABLE_PREORDER")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
