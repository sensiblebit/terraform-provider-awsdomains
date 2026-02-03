package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53domains"
	"github.com/aws/aws-sdk-go-v2/service/route53domains/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	tftypes "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &DomainRegistrationResource{}
var _ resource.ResourceWithImportState = &DomainRegistrationResource{}

type DomainRegistrationResource struct {
	client *route53domains.Client
}

type ContactModel struct {
	FirstName    tftypes.String `tfsdk:"first_name"`
	LastName     tftypes.String `tfsdk:"last_name"`
	Email        tftypes.String `tfsdk:"email"`
	PhoneNumber  tftypes.String `tfsdk:"phone_number"`
	AddressLine1 tftypes.String `tfsdk:"address_line_1"`
	AddressLine2 tftypes.String `tfsdk:"address_line_2"`
	City         tftypes.String `tfsdk:"city"`
	State        tftypes.String `tfsdk:"state"`
	ZipCode      tftypes.String `tfsdk:"zip_code"`
	CountryCode  tftypes.String `tfsdk:"country_code"`
	ContactType  tftypes.String `tfsdk:"contact_type"`
}

type DomainRegistrationResourceModel struct {
	ID                   tftypes.String   `tfsdk:"id"`
	DomainName           tftypes.String   `tfsdk:"domain_name"`
	DurationYears        tftypes.Int64    `tfsdk:"duration_years"`
	AutoRenew            tftypes.Bool     `tfsdk:"auto_renew"`
	AdminContact         *ContactModel    `tfsdk:"admin_contact"`
	RegistrantContact    *ContactModel    `tfsdk:"registrant_contact"`
	TechContact          *ContactModel    `tfsdk:"tech_contact"`
	AdminPrivacy         tftypes.Bool     `tfsdk:"admin_privacy"`
	RegistrantPrivacy    tftypes.Bool     `tfsdk:"registrant_privacy"`
	TechPrivacy          tftypes.Bool     `tfsdk:"tech_privacy"`
	Nameservers          []tftypes.String `tfsdk:"nameservers"`
	AllowDelete          tftypes.Bool     `tfsdk:"allow_delete"`
	Status               tftypes.String   `tfsdk:"status"`
	ExpirationDate       tftypes.String   `tfsdk:"expiration_date"`
	CreationDate         tftypes.String   `tfsdk:"creation_date"`
	RegistrationTimeout  tftypes.Int64    `tfsdk:"registration_timeout"`
}

func NewDomainRegistrationResource() resource.Resource {
	return &DomainRegistrationResource{}
}

func (r *DomainRegistrationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func contactSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Required:    true,
		Description: "Contact information for domain registration.",
		Attributes: map[string]schema.Attribute{
			"first_name": schema.StringAttribute{
				Required:    true,
				Description: "First name of the contact.",
			},
			"last_name": schema.StringAttribute{
				Required:    true,
				Description: "Last name of the contact.",
			},
			"email": schema.StringAttribute{
				Required:    true,
				Description: "Email address of the contact.",
			},
			"phone_number": schema.StringAttribute{
				Required:    true,
				Description: "Phone number in E.164 format (e.g., +1.5551234567).",
			},
			"address_line_1": schema.StringAttribute{
				Required:    true,
				Description: "First line of the street address.",
			},
			"address_line_2": schema.StringAttribute{
				Optional:    true,
				Description: "Second line of the street address.",
			},
			"city": schema.StringAttribute{
				Required:    true,
				Description: "City name.",
			},
			"state": schema.StringAttribute{
				Required:    true,
				Description: "State or province.",
			},
			"zip_code": schema.StringAttribute{
				Required:    true,
				Description: "Postal/ZIP code.",
			},
			"country_code": schema.StringAttribute{
				Required:    true,
				Description: "Two-letter country code (e.g., US).",
			},
			"contact_type": schema.StringAttribute{
				Optional:    true,
				Description: "Contact type: PERSON, COMPANY, ASSOCIATION, PUBLIC_BODY, or RESELLER.",
			},
		},
	}
}

func (r *DomainRegistrationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Registers and manages an AWS Route53 domain. By default, destroying this resource only removes it from Terraform state without deleting the actual domain. Set allow_delete = true to enable actual domain deletion on destroy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The domain name (used as the resource ID).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain_name": schema.StringAttribute{
				Required:    true,
				Description: "The domain name to register.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"duration_years": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1),
				Description: "Number of years to register the domain for (1-10).",
			},
			"auto_renew": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether to automatically renew the domain.",
			},
			"admin_contact":      contactSchema(),
			"registrant_contact": contactSchema(),
			"tech_contact":       contactSchema(),
			"admin_privacy": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Enable WHOIS privacy for admin contact.",
			},
			"registrant_privacy": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Enable WHOIS privacy for registrant contact.",
			},
			"tech_privacy": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Enable WHOIS privacy for tech contact.",
			},
			"nameservers": schema.ListAttribute{
				Optional:    true,
				ElementType: tftypes.StringType,
				Description: "List of nameserver hostnames for the domain.",
			},
			"allow_delete": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "DANGER: If true, destroying this resource will attempt to delete the domain registration. Default is false (domain is only removed from state).",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Current status of the domain.",
			},
			"expiration_date": schema.StringAttribute{
				Computed:    true,
				Description: "Expiration date of the domain registration.",
			},
			"creation_date": schema.StringAttribute{
				Computed:    true,
				Description: "Creation date of the domain registration.",
			},
			"registration_timeout": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(900),
				Description: "Timeout in seconds to wait for domain registration to complete (default: 900 = 15 minutes).",
			},
		},
	}
}

func (r *DomainRegistrationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*route53domains.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *route53domains.Client, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

func contactModelToAWS(m *ContactModel) *types.ContactDetail {
	if m == nil {
		return nil
	}

	contact := &types.ContactDetail{
		FirstName:    aws.String(m.FirstName.ValueString()),
		LastName:     aws.String(m.LastName.ValueString()),
		Email:        aws.String(m.Email.ValueString()),
		PhoneNumber:  aws.String(m.PhoneNumber.ValueString()),
		AddressLine1: aws.String(m.AddressLine1.ValueString()),
		City:         aws.String(m.City.ValueString()),
		State:        aws.String(m.State.ValueString()),
		ZipCode:      aws.String(m.ZipCode.ValueString()),
		CountryCode:  types.CountryCode(m.CountryCode.ValueString()),
	}

	if !m.AddressLine2.IsNull() && !m.AddressLine2.IsUnknown() {
		contact.AddressLine2 = aws.String(m.AddressLine2.ValueString())
	}

	if !m.ContactType.IsNull() && !m.ContactType.IsUnknown() {
		contact.ContactType = types.ContactType(m.ContactType.ValueString())
	} else {
		contact.ContactType = types.ContactTypePerson
	}

	return contact
}

func (r *DomainRegistrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DomainRegistrationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainName := data.DomainName.ValueString()
	tflog.Info(ctx, "Registering domain", map[string]interface{}{
		"domain": domainName,
	})

	// Build registration request
	registerInput := &route53domains.RegisterDomainInput{
		DomainName:        aws.String(domainName),
		DurationInYears:   aws.Int32(int32(data.DurationYears.ValueInt64())),
		AutoRenew:         aws.Bool(data.AutoRenew.ValueBool()),
		AdminContact:      contactModelToAWS(data.AdminContact),
		RegistrantContact: contactModelToAWS(data.RegistrantContact),
		TechContact:       contactModelToAWS(data.TechContact),
		PrivacyProtectAdminContact:      aws.Bool(data.AdminPrivacy.ValueBool()),
		PrivacyProtectRegistrantContact: aws.Bool(data.RegistrantPrivacy.ValueBool()),
		PrivacyProtectTechContact:       aws.Bool(data.TechPrivacy.ValueBool()),
	}

	// Register the domain
	registerOutput, err := r.client.RegisterDomain(ctx, registerInput)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error registering domain",
			fmt.Sprintf("Could not register domain %s: %s", domainName, err.Error()),
		)
		return
	}

	tflog.Info(ctx, "Domain registration initiated", map[string]interface{}{
		"domain":       domainName,
		"operation_id": *registerOutput.OperationId,
	})

	// Wait for registration to complete
	timeout := time.Duration(data.RegistrationTimeout.ValueInt64()) * time.Second
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		opDetail, err := r.client.GetOperationDetail(ctx, &route53domains.GetOperationDetailInput{
			OperationId: registerOutput.OperationId,
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error checking registration status",
				fmt.Sprintf("Could not check registration status for %s: %s", domainName, err.Error()),
			)
			return
		}

		tflog.Debug(ctx, "Registration operation status", map[string]interface{}{
			"domain": domainName,
			"status": opDetail.Status,
		})

		if opDetail.Status == types.OperationStatusSuccessful {
			break
		}
		if opDetail.Status == types.OperationStatusFailed {
			resp.Diagnostics.AddError(
				"Domain registration failed",
				fmt.Sprintf("Domain registration for %s failed: %s", domainName, aws.ToString(opDetail.Message)),
			)
			return
		}
		if opDetail.Status == types.OperationStatusError {
			resp.Diagnostics.AddError(
				"Domain registration error",
				fmt.Sprintf("Domain registration for %s encountered an error: %s", domainName, aws.ToString(opDetail.Message)),
			)
			return
		}

		time.Sleep(10 * time.Second)
	}

	// Update nameservers if specified
	if len(data.Nameservers) > 0 {
		var nameservers []types.Nameserver
		for _, ns := range data.Nameservers {
			nameservers = append(nameservers, types.Nameserver{
				Name: aws.String(ns.ValueString()),
			})
		}

		_, err := r.client.UpdateDomainNameservers(ctx, &route53domains.UpdateDomainNameserversInput{
			DomainName:  aws.String(domainName),
			Nameservers: nameservers,
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating nameservers",
				fmt.Sprintf("Could not update nameservers for %s: %s", domainName, err.Error()),
			)
			return
		}
	}

	// Get domain details
	domainDetail, err := r.client.GetDomainDetail(ctx, &route53domains.GetDomainDetailInput{
		DomainName: aws.String(domainName),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading domain details",
			fmt.Sprintf("Could not read domain details for %s: %s", domainName, err.Error()),
		)
		return
	}

	// Update state
	data.ID = tftypes.StringValue(domainName)
	if domainDetail.ExpirationDate != nil {
		data.ExpirationDate = tftypes.StringValue(domainDetail.ExpirationDate.Format(time.RFC3339))
	}
	if domainDetail.CreationDate != nil {
		data.CreationDate = tftypes.StringValue(domainDetail.CreationDate.Format(time.RFC3339))
	}
	if len(domainDetail.StatusList) > 0 {
		data.Status = tftypes.StringValue(string(domainDetail.StatusList[0]))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainRegistrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DomainRegistrationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainName := data.DomainName.ValueString()

	domainDetail, err := r.client.GetDomainDetail(ctx, &route53domains.GetDomainDetailInput{
		DomainName: aws.String(domainName),
	})
	if err != nil {
		// If domain not found, remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	// Update computed fields
	data.ID = tftypes.StringValue(domainName)
	if domainDetail.AutoRenew != nil {
		data.AutoRenew = tftypes.BoolValue(*domainDetail.AutoRenew)
	}
	if domainDetail.ExpirationDate != nil {
		data.ExpirationDate = tftypes.StringValue(domainDetail.ExpirationDate.Format(time.RFC3339))
	}
	if domainDetail.CreationDate != nil {
		data.CreationDate = tftypes.StringValue(domainDetail.CreationDate.Format(time.RFC3339))
	}
	if len(domainDetail.StatusList) > 0 {
		data.Status = tftypes.StringValue(string(domainDetail.StatusList[0]))
	}

	// Update nameservers from AWS
	if len(domainDetail.Nameservers) > 0 {
		var nameservers []tftypes.String
		for _, ns := range domainDetail.Nameservers {
			nameservers = append(nameservers, tftypes.StringValue(aws.ToString(ns.Name)))
		}
		data.Nameservers = nameservers
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainRegistrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DomainRegistrationResourceModel
	var state DomainRegistrationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainName := data.DomainName.ValueString()

	// Update auto-renew if changed
	if data.AutoRenew.ValueBool() != state.AutoRenew.ValueBool() {
		if data.AutoRenew.ValueBool() {
			_, err := r.client.EnableDomainAutoRenew(ctx, &route53domains.EnableDomainAutoRenewInput{
				DomainName: aws.String(domainName),
			})
			if err != nil {
				resp.Diagnostics.AddError(
					"Error enabling auto-renew",
					fmt.Sprintf("Could not enable auto-renew for %s: %s", domainName, err.Error()),
				)
				return
			}
		} else {
			_, err := r.client.DisableDomainAutoRenew(ctx, &route53domains.DisableDomainAutoRenewInput{
				DomainName: aws.String(domainName),
			})
			if err != nil {
				resp.Diagnostics.AddError(
					"Error disabling auto-renew",
					fmt.Sprintf("Could not disable auto-renew for %s: %s", domainName, err.Error()),
				)
				return
			}
		}
	}

	// Update nameservers if changed
	if len(data.Nameservers) > 0 {
		var nameservers []types.Nameserver
		for _, ns := range data.Nameservers {
			nameservers = append(nameservers, types.Nameserver{
				Name: aws.String(ns.ValueString()),
			})
		}

		_, err := r.client.UpdateDomainNameservers(ctx, &route53domains.UpdateDomainNameserversInput{
			DomainName:  aws.String(domainName),
			Nameservers: nameservers,
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating nameservers",
				fmt.Sprintf("Could not update nameservers for %s: %s", domainName, err.Error()),
			)
			return
		}
	}

	// Update contacts if changed
	_, err := r.client.UpdateDomainContact(ctx, &route53domains.UpdateDomainContactInput{
		DomainName:        aws.String(domainName),
		AdminContact:      contactModelToAWS(data.AdminContact),
		RegistrantContact: contactModelToAWS(data.RegistrantContact),
		TechContact:       contactModelToAWS(data.TechContact),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating contacts",
			fmt.Sprintf("Could not update contacts for %s: %s", domainName, err.Error()),
		)
		return
	}

	// Update privacy settings
	_, err = r.client.UpdateDomainContactPrivacy(ctx, &route53domains.UpdateDomainContactPrivacyInput{
		DomainName:      aws.String(domainName),
		AdminPrivacy:    aws.Bool(data.AdminPrivacy.ValueBool()),
		RegistrantPrivacy: aws.Bool(data.RegistrantPrivacy.ValueBool()),
		TechPrivacy:     aws.Bool(data.TechPrivacy.ValueBool()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating privacy settings",
			fmt.Sprintf("Could not update privacy settings for %s: %s", domainName, err.Error()),
		)
		return
	}

	// Refresh state
	domainDetail, err := r.client.GetDomainDetail(ctx, &route53domains.GetDomainDetailInput{
		DomainName: aws.String(domainName),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading domain details",
			fmt.Sprintf("Could not read domain details for %s: %s", domainName, err.Error()),
		)
		return
	}

	data.ID = tftypes.StringValue(domainName)
	if domainDetail.ExpirationDate != nil {
		data.ExpirationDate = tftypes.StringValue(domainDetail.ExpirationDate.Format(time.RFC3339))
	}
	if domainDetail.CreationDate != nil {
		data.CreationDate = tftypes.StringValue(domainDetail.CreationDate.Format(time.RFC3339))
	}
	if len(domainDetail.StatusList) > 0 {
		data.Status = tftypes.StringValue(string(domainDetail.StatusList[0]))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainRegistrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DomainRegistrationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainName := data.DomainName.ValueString()

	// Check if deletion is allowed
	if !data.AllowDelete.ValueBool() {
		tflog.Warn(ctx, "Domain will be removed from state only (allow_delete = false)", map[string]interface{}{
			"domain": domainName,
		})
		// Just remove from state, don't actually delete
		return
	}

	tflog.Warn(ctx, "DELETING DOMAIN REGISTRATION (allow_delete = true)", map[string]interface{}{
		"domain": domainName,
	})

	// Attempt to delete the domain
	_, err := r.client.DeleteDomain(ctx, &route53domains.DeleteDomainInput{
		DomainName: aws.String(domainName),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting domain",
			fmt.Sprintf("Could not delete domain %s: %s. Note: Domain deletion may not be supported by the registry. The domain has been removed from Terraform state.", domainName, err.Error()),
		)
		// Still remove from state even if delete fails
		return
	}

	tflog.Info(ctx, "Domain deletion initiated", map[string]interface{}{
		"domain": domainName,
	})
}

func (r *DomainRegistrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("domain_name"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
