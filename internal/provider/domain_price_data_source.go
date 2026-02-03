package provider

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/route53domains"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &DomainPriceDataSource{}

type DomainPriceDataSource struct {
	client *route53domains.Client
}

type DomainPriceDataSourceModel struct {
	ID                  types.String  `tfsdk:"id"`
	TLD                 types.String  `tfsdk:"tld"`
	RegistrationPrice   types.Float64 `tfsdk:"registration_price"`
	RenewalPrice        types.Float64 `tfsdk:"renewal_price"`
	TransferPrice       types.Float64 `tfsdk:"transfer_price"`
	ChangeOwnershipPrice types.Float64 `tfsdk:"change_ownership_price"`
	RestorationPrice    types.Float64 `tfsdk:"restoration_price"`
	Currency            types.String  `tfsdk:"currency"`
}

func NewDomainPriceDataSource() datasource.DataSource {
	return &DomainPriceDataSource{}
}

func (d *DomainPriceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_price"
}

func (d *DomainPriceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Get pricing information for a domain TLD.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The TLD.",
			},
			"tld": schema.StringAttribute{
				Required:    true,
				Description: "The top-level domain (e.g., 'com', 'net', 'org').",
			},
			"registration_price": schema.Float64Attribute{
				Computed:    true,
				Description: "Price to register a new domain.",
			},
			"renewal_price": schema.Float64Attribute{
				Computed:    true,
				Description: "Price to renew a domain.",
			},
			"transfer_price": schema.Float64Attribute{
				Computed:    true,
				Description: "Price to transfer a domain.",
			},
			"change_ownership_price": schema.Float64Attribute{
				Computed:    true,
				Description: "Price to change domain ownership.",
			},
			"restoration_price": schema.Float64Attribute{
				Computed:    true,
				Description: "Price to restore a deleted domain.",
			},
			"currency": schema.StringAttribute{
				Computed:    true,
				Description: "Currency code (e.g., USD).",
			},
		},
	}
}

func (d *DomainPriceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DomainPriceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DomainPriceDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tld := data.TLD.ValueString()

	// List prices and find the one for our TLD
	paginator := route53domains.NewListPricesPaginator(d.client, &route53domains.ListPricesInput{
		Tld: &tld,
	})

	var found bool
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error listing domain prices",
				fmt.Sprintf("Could not list prices for TLD %s: %s", tld, err.Error()),
			)
			return
		}

		for _, price := range page.Prices {
			if price.Name != nil && *price.Name == tld {
				found = true
				data.ID = types.StringValue(tld)

				if price.RegistrationPrice != nil {
					data.RegistrationPrice = types.Float64Value(price.RegistrationPrice.Price)
					data.Currency = types.StringValue(*price.RegistrationPrice.Currency)
				}
				if price.RenewalPrice != nil {
					data.RenewalPrice = types.Float64Value(price.RenewalPrice.Price)
				}
				if price.TransferPrice != nil {
					data.TransferPrice = types.Float64Value(price.TransferPrice.Price)
				}
				if price.ChangeOwnershipPrice != nil {
					data.ChangeOwnershipPrice = types.Float64Value(price.ChangeOwnershipPrice.Price)
				}
				if price.RestorationPrice != nil {
					data.RestorationPrice = types.Float64Value(price.RestorationPrice.Price)
				}
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		resp.Diagnostics.AddError(
			"TLD not found",
			fmt.Sprintf("No pricing information found for TLD: %s", tld),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
