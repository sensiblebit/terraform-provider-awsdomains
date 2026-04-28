package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	route53types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/aws/aws-sdk-go-v2/service/route53domains"
	"github.com/aws/aws-sdk-go-v2/service/route53domains/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	tftypes "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &DomainRegistrationResource{}
var _ resource.ResourceWithImportState = &DomainRegistrationResource{}
var _ resource.ResourceWithModifyPlan = &DomainRegistrationResource{}
var _ resource.ResourceWithValidateConfig = &DomainRegistrationResource{}

// DomainRegistrationResource manages Route53 Domains registrations.
type DomainRegistrationResource struct {
	client        route53DomainsAPI
	route53Client route53API
	defaultTags   map[string]string
}

type route53API interface {
	ListHostedZonesByName(ctx context.Context, params *route53.ListHostedZonesByNameInput, optFns ...func(*route53.Options)) (*route53.ListHostedZonesByNameOutput, error)
	ListResourceRecordSets(ctx context.Context, params *route53.ListResourceRecordSetsInput, optFns ...func(*route53.Options)) (*route53.ListResourceRecordSetsOutput, error)
	DeleteHostedZone(ctx context.Context, params *route53.DeleteHostedZoneInput, optFns ...func(*route53.Options)) (*route53.DeleteHostedZoneOutput, error)
}

type route53DomainsAPI interface {
	GetDomainDetail(ctx context.Context, params *route53domains.GetDomainDetailInput, optFns ...func(*route53domains.Options)) (*route53domains.GetDomainDetailOutput, error)
	RegisterDomain(ctx context.Context, params *route53domains.RegisterDomainInput, optFns ...func(*route53domains.Options)) (*route53domains.RegisterDomainOutput, error)
	GetOperationDetail(ctx context.Context, params *route53domains.GetOperationDetailInput, optFns ...func(*route53domains.Options)) (*route53domains.GetOperationDetailOutput, error)
	UpdateDomainNameservers(ctx context.Context, params *route53domains.UpdateDomainNameserversInput, optFns ...func(*route53domains.Options)) (*route53domains.UpdateDomainNameserversOutput, error)
	EnableDomainAutoRenew(ctx context.Context, params *route53domains.EnableDomainAutoRenewInput, optFns ...func(*route53domains.Options)) (*route53domains.EnableDomainAutoRenewOutput, error)
	DisableDomainAutoRenew(ctx context.Context, params *route53domains.DisableDomainAutoRenewInput, optFns ...func(*route53domains.Options)) (*route53domains.DisableDomainAutoRenewOutput, error)
	UpdateDomainContact(ctx context.Context, params *route53domains.UpdateDomainContactInput, optFns ...func(*route53domains.Options)) (*route53domains.UpdateDomainContactOutput, error)
	UpdateDomainContactPrivacy(ctx context.Context, params *route53domains.UpdateDomainContactPrivacyInput, optFns ...func(*route53domains.Options)) (*route53domains.UpdateDomainContactPrivacyOutput, error)
	DeleteDomain(ctx context.Context, params *route53domains.DeleteDomainInput, optFns ...func(*route53domains.Options)) (*route53domains.DeleteDomainOutput, error)
	ListTagsForDomain(ctx context.Context, params *route53domains.ListTagsForDomainInput, optFns ...func(*route53domains.Options)) (*route53domains.ListTagsForDomainOutput, error)
	UpdateTagsForDomain(ctx context.Context, params *route53domains.UpdateTagsForDomainInput, optFns ...func(*route53domains.Options)) (*route53domains.UpdateTagsForDomainOutput, error)
	DeleteTagsForDomain(ctx context.Context, params *route53domains.DeleteTagsForDomainInput, optFns ...func(*route53domains.Options)) (*route53domains.DeleteTagsForDomainOutput, error)
}

const (
	registrarHostedZoneComment               = "HostedZone created by Route53 Registrar"
	registrarHostedZoneCallerReferencePrefix = "RISWorkflow-RD:"
)

var (
	errRegistrarHostedZoneNotFound        = errors.New("registrar-created hosted zone not found")
	errMultipleRegistrarHostedZones       = errors.New("multiple registrar-created hosted zones found")
	errRegistrarHostedZoneCommentMismatch = errors.New("hosted zone comment does not match expected registrar comment")
	errHostedZoneHasCustomRecord          = errors.New("hosted zone has custom record, not deleting")
)

// ContactModel stores the contact fields used for domain registration.
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

// DomainRegistrationResourceModel stores Terraform state for the domain resource.
type DomainRegistrationResourceModel struct {
	ID                  tftypes.String `tfsdk:"id"`
	DomainName          tftypes.String `tfsdk:"domain_name"`
	DurationYears       tftypes.Int64  `tfsdk:"duration_years"`
	AutoRenew           tftypes.Bool   `tfsdk:"auto_renew"`
	AdminContact        *ContactModel  `tfsdk:"admin_contact"`
	RegistrantContact   *ContactModel  `tfsdk:"registrant_contact"`
	TechContact         *ContactModel  `tfsdk:"tech_contact"`
	AdminPrivacy        tftypes.Bool   `tfsdk:"admin_privacy"`
	RegistrantPrivacy   tftypes.Bool   `tfsdk:"registrant_privacy"`
	TechPrivacy         tftypes.Bool   `tfsdk:"tech_privacy"`
	Nameservers         tftypes.List   `tfsdk:"nameservers"`
	Tags                tftypes.Map    `tfsdk:"tags"`
	TagsAll             tftypes.Map    `tfsdk:"tags_all"`
	AllowDelete         tftypes.Bool   `tfsdk:"allow_delete"`
	DeleteHostedZone    tftypes.Bool   `tfsdk:"delete_hosted_zone"`
	Status              tftypes.String `tfsdk:"status"`
	ExpirationDate      tftypes.String `tfsdk:"expiration_date"`
	CreationDate        tftypes.String `tfsdk:"creation_date"`
	RegistrationTimeout tftypes.Int64  `tfsdk:"registration_timeout"`
	HostedZoneID        tftypes.String `tfsdk:"hosted_zone_id"`
}

// NewDomainRegistrationResource creates the domain registration resource.
func NewDomainRegistrationResource() resource.Resource {
	return &DomainRegistrationResource{}
}

// Metadata sets the resource type name.
func (r *DomainRegistrationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
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

// Schema describes the resource attributes.
func (r *DomainRegistrationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"tags": schema.MapAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: tftypes.StringType,
				Default:     mapdefault.StaticValue(emptyFrameworkStringMap()),
				Description: "Resource-level tags for the domain. These override provider default_tags with the same key.",
			},
			"tags_all": schema.MapAttribute{
				Computed:    true,
				ElementType: tftypes.StringType,
				Description: "All tags applied to the domain, including provider default_tags and resource-level tags.",
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
			},
			"allow_delete": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "DANGER: If true, destroying this resource will attempt to delete the domain registration. Default is false (domain is only removed from state).",
			},
			"delete_hosted_zone": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Delete the auto-created Route53 hosted zone after domain registration. Use when pointing to external DNS. Only deletes if zone is public, has registrar comment, and contains only NS/SOA records.",
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
			"hosted_zone_id": schema.StringAttribute{
				Computed:    true,
				Description: "The Route53 hosted zone ID automatically created for this domain.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// ValidateConfig validates resource configuration before planning or applying changes.
func (r *DomainRegistrationResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data DomainRegistrationResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	validateNameserversConfig(&resp.Diagnostics, data.Nameservers)

	if !frameworkMapElementsKnown(data.Tags) {
		return
	}

	resourceTags, diags := frameworkMapToStringMap(ctx, data.Tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	addTagValidationDiagnostics(&resp.Diagnostics, path.Root("tags"), resourceTags)
}

// Configure loads shared provider clients into the resource.
func (r *DomainRegistrationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*providerData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *providerData, got: %T", req.ProviderData),
		)
		return
	}

	r.client = providerData.DomainsClient
	r.route53Client = providerData.Route53Client
	r.defaultTags = cloneTags(providerData.DefaultTags)
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

func awsStringValue(value *string) tftypes.String {
	if value == nil {
		return tftypes.StringNull()
	}
	return tftypes.StringValue(*value)
}

func awsOptionalStringValue(value *string, prior tftypes.String) tftypes.String {
	if value == nil {
		return tftypes.StringNull()
	}
	if prior.IsNull() && *value == "" {
		return tftypes.StringNull()
	}
	return tftypes.StringValue(*value)
}

func awsContactTypeValue(value types.ContactType, prior tftypes.String) tftypes.String {
	if value == "" {
		return tftypes.StringNull()
	}
	if prior.IsNull() && value == types.ContactTypePerson {
		return tftypes.StringNull()
	}
	return tftypes.StringValue(string(value))
}

func awsCountryCodeValue(value types.CountryCode) tftypes.String {
	if value == "" {
		return tftypes.StringNull()
	}
	return tftypes.StringValue(string(value))
}

func awsContactToModel(contact *types.ContactDetail, prior *ContactModel) *ContactModel {
	if contact == nil {
		return prior
	}

	priorAddressLine2 := tftypes.StringNull()
	priorContactType := tftypes.StringNull()
	if prior != nil {
		priorAddressLine2 = prior.AddressLine2
		priorContactType = prior.ContactType
	}

	return &ContactModel{
		FirstName:    awsStringValue(contact.FirstName),
		LastName:     awsStringValue(contact.LastName),
		Email:        awsStringValue(contact.Email),
		PhoneNumber:  awsStringValue(contact.PhoneNumber),
		AddressLine1: awsStringValue(contact.AddressLine1),
		AddressLine2: awsOptionalStringValue(contact.AddressLine2, priorAddressLine2),
		City:         awsStringValue(contact.City),
		State:        awsStringValue(contact.State),
		ZipCode:      awsStringValue(contact.ZipCode),
		CountryCode:  awsCountryCodeValue(contact.CountryCode),
		ContactType:  awsContactTypeValue(contact.ContactType, priorContactType),
	}
}

func contactModelsEqual(a, b *ContactModel) bool {
	if a == nil || b == nil {
		return a == b
	}

	return a.FirstName.Equal(b.FirstName) &&
		a.LastName.Equal(b.LastName) &&
		a.Email.Equal(b.Email) &&
		a.PhoneNumber.Equal(b.PhoneNumber) &&
		a.AddressLine1.Equal(b.AddressLine1) &&
		a.AddressLine2.Equal(b.AddressLine2) &&
		a.City.Equal(b.City) &&
		a.State.Equal(b.State) &&
		a.ZipCode.Equal(b.ZipCode) &&
		a.CountryCode.Equal(b.CountryCode) &&
		a.ContactType.Equal(b.ContactType)
}

func domainContactsEqual(a, b DomainRegistrationResourceModel) bool {
	return contactModelsEqual(a.AdminContact, b.AdminContact) &&
		contactModelsEqual(a.RegistrantContact, b.RegistrantContact) &&
		contactModelsEqual(a.TechContact, b.TechContact)
}

func domainPrivacySettingsEqual(a, b DomainRegistrationResourceModel) bool {
	return a.AdminPrivacy.Equal(b.AdminPrivacy) &&
		a.RegistrantPrivacy.Equal(b.RegistrantPrivacy) &&
		a.TechPrivacy.Equal(b.TechPrivacy)
}

func frameworkListToAWSNameservers(ctx context.Context, value tftypes.List) ([]types.Nameserver, diag.Diagnostics) {
	var nameserverNames []string
	if value.IsNull() || value.IsUnknown() {
		return nil, nil
	}

	diags := value.ElementsAs(ctx, &nameserverNames, false)
	if diags.HasError() {
		return nil, diags
	}

	nameservers := make([]types.Nameserver, 0, len(nameserverNames))
	for _, nameserverName := range nameserverNames {
		nameservers = append(nameservers, types.Nameserver{
			Name: aws.String(nameserverName),
		})
	}

	return nameservers, diags
}

func stringSliceToFrameworkList(ctx context.Context, values []string) (tftypes.List, diag.Diagnostics) {
	return tftypes.ListValueFrom(ctx, tftypes.StringType, values)
}

func validateNameserversConfig(diags *diag.Diagnostics, value tftypes.List) {
	if value.IsNull() || value.IsUnknown() {
		return
	}
	if len(value.Elements()) > 0 {
		return
	}

	diags.AddAttributeError(
		path.Root("nameservers"),
		"Invalid Nameservers Configuration",
		"Set nameservers to at least one hostname or omit the attribute. Clearing nameservers is not supported by Route53 Domains and would leave Terraform state inconsistent with AWS.",
	)
}

func nameserversRemoved(plan, state tftypes.List) bool {
	if plan.IsUnknown() || state.IsNull() || state.IsUnknown() || len(state.Elements()) == 0 {
		return false
	}
	if plan.IsNull() {
		return true
	}
	return len(plan.Elements()) == 0
}

func (r *DomainRegistrationResource) populateDomainDetailState(ctx context.Context, data *DomainRegistrationResourceModel, domainDetail *route53domains.GetDomainDetailOutput) diag.Diagnostics {
	var diags diag.Diagnostics

	domainName := data.DomainName.ValueString()
	if domainDetail.DomainName != nil {
		domainName = aws.ToString(domainDetail.DomainName)
		data.DomainName = tftypes.StringValue(domainName)
	}
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
		data.Status = tftypes.StringValue(domainDetail.StatusList[0])
	}

	data.AdminContact = awsContactToModel(domainDetail.AdminContact, data.AdminContact)
	data.RegistrantContact = awsContactToModel(domainDetail.RegistrantContact, data.RegistrantContact)
	data.TechContact = awsContactToModel(domainDetail.TechContact, data.TechContact)

	if domainDetail.AdminPrivacy != nil {
		data.AdminPrivacy = tftypes.BoolValue(*domainDetail.AdminPrivacy)
	}
	if domainDetail.RegistrantPrivacy != nil {
		data.RegistrantPrivacy = tftypes.BoolValue(*domainDetail.RegistrantPrivacy)
	}
	if domainDetail.TechPrivacy != nil {
		data.TechPrivacy = tftypes.BoolValue(*domainDetail.TechPrivacy)
	}

	if len(domainDetail.Nameservers) > 0 {
		nameservers := make([]string, 0, len(domainDetail.Nameservers))
		for _, ns := range domainDetail.Nameservers {
			nameservers = append(nameservers, aws.ToString(ns.Name))
		}
		nameserversList, nameserverDiags := stringSliceToFrameworkList(ctx, nameservers)
		diags.Append(nameserverDiags...)
		if diags.HasError() {
			return diags
		}
		data.Nameservers = nameserversList
	}

	return diags
}

func isExactHostedZoneName(zoneName, domainName string) bool {
	return strings.TrimSuffix(zoneName, ".") == domainName
}

func selectRegistrarHostedZone(domainName string, zones []route53types.HostedZone) (*route53types.HostedZone, error) {
	var matches []route53types.HostedZone

	for _, zone := range zones {
		if !isExactHostedZoneName(aws.ToString(zone.Name), domainName) {
			continue
		}
		if zone.Config != nil && zone.Config.PrivateZone {
			continue
		}
		if !strings.HasPrefix(aws.ToString(zone.CallerReference), registrarHostedZoneCallerReferencePrefix) {
			continue
		}

		matches = append(matches, zone)
	}

	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("%w: %s", errRegistrarHostedZoneNotFound, domainName)
	case 1:
		return &matches[0], nil
	default:
		return nil, fmt.Errorf("%w: %s", errMultipleRegistrarHostedZones, domainName)
	}
}

func (r *DomainRegistrationResource) listExactHostedZones(ctx context.Context, domainName string) ([]route53types.HostedZone, error) {
	input := &route53.ListHostedZonesByNameInput{
		DNSName:  aws.String(domainName),
		MaxItems: aws.Int32(100),
	}

	var zones []route53types.HostedZone

	for {
		output, err := r.route53Client.ListHostedZonesByName(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to list hosted zones: %w", err)
		}

		for _, zone := range output.HostedZones {
			zoneName := strings.TrimSuffix(aws.ToString(zone.Name), ".")
			if zoneName == domainName {
				zones = append(zones, zone)
				continue
			}

			return zones, nil
		}

		if !output.IsTruncated || output.NextDNSName == nil {
			return zones, nil
		}

		input = &route53.ListHostedZonesByNameInput{
			DNSName:      output.NextDNSName,
			HostedZoneId: output.NextHostedZoneId,
			MaxItems:     aws.Int32(100),
		}
	}
}

// findHostedZoneID looks up the registrar-created Route53 hosted zone ID for a domain
func (r *DomainRegistrationResource) findHostedZoneID(ctx context.Context, domainName string) (string, error) {
	zones, err := r.listExactHostedZones(ctx, domainName)
	if err != nil {
		return "", err
	}

	zone, err := selectRegistrarHostedZone(domainName, zones)
	if err != nil {
		return "", err
	}

	return strings.TrimPrefix(aws.ToString(zone.Id), "/hostedzone/"), nil
}

// deleteRegistrarHostedZone safely deletes the hosted zone only if ALL conditions are met:
// 1. Zone name matches the domain exactly
// 2. Zone is public (not private)
// 3. Zone comment is "HostedZone created by Route53 Registrar"
// 4. Zone contains only NS and SOA records (no custom records)
func (r *DomainRegistrationResource) deleteRegistrarHostedZone(ctx context.Context, domainName string) error {
	zones, err := r.listExactHostedZones(ctx, domainName)
	if err != nil {
		return err
	}

	zone, err := selectRegistrarHostedZone(domainName, zones)
	if err != nil {
		return err
	}

	zoneID := aws.ToString(zone.Id)

	comment := ""
	if zone.Config != nil && zone.Config.Comment != nil {
		comment = *zone.Config.Comment
	}
	if comment != registrarHostedZoneComment {
		tflog.Warn(ctx, "Hosted zone not created by Route53 Registrar, skipping deletion", map[string]any{
			"domain":  domainName,
			"zone_id": zoneID,
			"comment": comment,
		})
		return fmt.Errorf("%w: %q", errRegistrarHostedZoneCommentMismatch, comment)
	}

	listInput := &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(zoneID),
	}

	for {
		recordsOutput, err := r.route53Client.ListResourceRecordSets(ctx, listInput)
		if err != nil {
			return fmt.Errorf("failed to list records in hosted zone: %w", err)
		}

		for _, record := range recordsOutput.ResourceRecordSets {
			recordType := string(record.Type)
			if recordType != "NS" && recordType != "SOA" {
				tflog.Warn(ctx, "Hosted zone has custom records, skipping deletion", map[string]any{
					"domain":      domainName,
					"zone_id":     zoneID,
					"record_name": aws.ToString(record.Name),
					"record_type": recordType,
				})
				return fmt.Errorf("%w: %s %s", errHostedZoneHasCustomRecord, aws.ToString(record.Name), recordType)
			}
		}

		if !recordsOutput.IsTruncated {
			break
		}

		listInput.StartRecordName = recordsOutput.NextRecordName
		listInput.StartRecordType = recordsOutput.NextRecordType
		listInput.StartRecordIdentifier = recordsOutput.NextRecordIdentifier
	}

	tflog.Info(ctx, "Deleting Route53 Registrar hosted zone", map[string]any{
		"domain":  domainName,
		"zone_id": zoneID,
	})

	_, err = r.route53Client.DeleteHostedZone(ctx, &route53.DeleteHostedZoneInput{
		Id: aws.String(zoneID),
	})
	if err != nil {
		return fmt.Errorf("failed to delete hosted zone: %w", err)
	}

	return nil
}

// ModifyPlan computes tags_all from provider default_tags and resource-level tags.
func (r *DomainRegistrationResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	var data DomainRegistrationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !frameworkMapElementsKnown(data.Tags) {
		return
	}

	resourceTags, diags := frameworkMapToStringMap(ctx, data.Tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	mergedTags := mergeTags(r.defaultTags, resourceTags)
	addTagValidationDiagnostics(&resp.Diagnostics, path.Root("tags_all"), mergedTags)
	if resp.Diagnostics.HasError() {
		return
	}

	tagsAll, diags := stringMapToFrameworkMap(mergedTags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("tags_all"), tagsAll)...)
}

func (r *DomainRegistrationResource) listDomainTags(ctx context.Context, domainName string) (map[string]string, error) {
	output, err := r.client.ListTagsForDomain(ctx, &route53domains.ListTagsForDomainInput{
		DomainName: aws.String(domainName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list domain tags: %w", err)
	}

	return awsTagsToStringMap(output.TagList), nil
}

func (r *DomainRegistrationResource) syncDomainTags(ctx context.Context, domainName string, currentTags, desiredTags map[string]string) error {
	tagsToDelete := tagKeysToDelete(currentTags, desiredTags)
	if len(tagsToDelete) > 0 {
		_, err := r.client.DeleteTagsForDomain(ctx, &route53domains.DeleteTagsForDomainInput{
			DomainName:   aws.String(domainName),
			TagsToDelete: tagsToDelete,
		})
		if err != nil {
			return fmt.Errorf("failed to delete domain tags: %w", err)
		}
	}

	updatedTags := tagsToUpdate(currentTags, desiredTags)
	if len(updatedTags) > 0 {
		_, err := r.client.UpdateTagsForDomain(ctx, &route53domains.UpdateTagsForDomainInput{
			DomainName:   aws.String(domainName),
			TagsToUpdate: stringMapToAWSTags(updatedTags),
		})
		if err != nil {
			return fmt.Errorf("failed to update domain tags: %w", err)
		}
	}

	return nil
}

// Create registers the domain and records its hosted zone metadata.
func (r *DomainRegistrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DomainRegistrationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainName := data.DomainName.ValueString()
	tflog.Info(ctx, "Registering domain", map[string]any{
		"domain": domainName,
	})

	durationYears := data.DurationYears.ValueInt64()
	if durationYears < 1 || durationYears > 10 {
		resp.Diagnostics.AddError(
			"Invalid duration_years value",
			fmt.Sprintf("Expected duration_years to be between 1 and 10, got %d", durationYears),
		)
		return
	}

	resourceTags, diags := frameworkMapToStringMap(ctx, data.Tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	desiredTags := mergeTags(r.defaultTags, resourceTags)
	addTagValidationDiagnostics(&resp.Diagnostics, path.Root("tags_all"), desiredTags)
	validateNameserversConfig(&resp.Diagnostics, data.Nameservers)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build registration request
	registerInput := &route53domains.RegisterDomainInput{
		DomainName:                      aws.String(domainName),
		DurationInYears:                 aws.Int32(int32(durationYears)),
		AutoRenew:                       aws.Bool(data.AutoRenew.ValueBool()),
		AdminContact:                    contactModelToAWS(data.AdminContact),
		RegistrantContact:               contactModelToAWS(data.RegistrantContact),
		TechContact:                     contactModelToAWS(data.TechContact),
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

	tflog.Info(ctx, "Domain registration initiated", map[string]any{
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

		tflog.Debug(ctx, "Registration operation status", map[string]any{
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

	tagsAllSource := desiredTags
	if len(desiredTags) > 0 {
		err = r.syncDomainTags(ctx, domainName, map[string]string{}, desiredTags)
		if err != nil {
			resp.Diagnostics.AddWarning(
				"Domain Registered Without Tags",
				fmt.Sprintf("The domain %s was registered, but tags could not be updated: %s. The resource will remain in Terraform state so a later apply can retry tag reconciliation.", domainName, err.Error()),
			)
			if remoteTags, listErr := r.listDomainTags(ctx, domainName); listErr == nil {
				tagsAllSource = remoteTags
			}
		}
	}

	nameservers, diags := frameworkListToAWSNameservers(ctx, data.Nameservers)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if len(nameservers) > 0 {
		_, err := r.client.UpdateDomainNameservers(ctx, &route53domains.UpdateDomainNameserversInput{
			DomainName:  aws.String(domainName),
			Nameservers: nameservers,
		})
		if err != nil {
			resp.Diagnostics.AddWarning(
				"Domain Registered Without Nameserver Update",
				fmt.Sprintf("The domain %s was registered, but nameservers could not be updated: %s. The resource will remain in Terraform state so a later apply can retry nameserver reconciliation.", domainName, err.Error()),
			)
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
	resp.Diagnostics.Append(r.populateDomainDetailState(ctx, &data, domainDetail)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tagsAll, diags := stringMapToFrameworkMap(tagsAllSource)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.TagsAll = tagsAll

	// Handle the auto-created hosted zone
	if data.DeleteHostedZone.ValueBool() {
		// Delete the registrar-created hosted zone
		err := r.deleteRegistrarHostedZone(ctx, domainName)
		if err != nil {
			tflog.Warn(ctx, "Could not delete hosted zone", map[string]any{
				"domain": domainName,
				"error":  err.Error(),
			})
			// Still try to get the zone ID for state
			if hostedZoneID, lookupErr := r.findHostedZoneID(ctx, domainName); lookupErr == nil {
				data.HostedZoneID = tftypes.StringValue(hostedZoneID)
			} else {
				data.HostedZoneID = tftypes.StringNull()
			}
		} else {
			tflog.Info(ctx, "Deleted auto-created hosted zone", map[string]any{
				"domain": domainName,
			})
			data.HostedZoneID = tftypes.StringNull()
		}
	} else {
		// Look up the auto-created hosted zone
		hostedZoneID, err := r.findHostedZoneID(ctx, domainName)
		if err != nil {
			tflog.Warn(ctx, "Could not find hosted zone for domain", map[string]any{
				"domain": domainName,
				"error":  err.Error(),
			})
			data.HostedZoneID = tftypes.StringNull()
		} else {
			data.HostedZoneID = tftypes.StringValue(hostedZoneID)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the resource state from Route53 Domains.
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

	resp.Diagnostics.Append(r.populateDomainDetailState(ctx, &data, domainDetail)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if tagManagementEnabled(r.defaultTags, data.Tags, data.TagsAll) {
		remoteTags, err := r.listDomainTags(ctx, domainName)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error reading domain tags",
				fmt.Sprintf("Could not read tags for %s: %s", domainName, err.Error()),
			)
			return
		}

		priorResourceTags, diags := frameworkMapToStringMap(ctx, data.Tags)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.Tags, diags = stringMapToFrameworkMap(resourceTagsFromRemote(remoteTags, priorResourceTags))
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.TagsAll, diags = stringMapToFrameworkMap(remoteTags)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		data.Tags = emptyFrameworkStringMap()
		data.TagsAll = emptyFrameworkStringMap()
	}

	// Refresh hosted zone ID
	hostedZoneID, err := r.findHostedZoneID(ctx, domainName)
	if err != nil {
		data.HostedZoneID = tftypes.StringNull()
	} else {
		data.HostedZoneID = tftypes.StringValue(hostedZoneID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies mutable domain settings and refreshes state.
func (r *DomainRegistrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DomainRegistrationResourceModel
	var state DomainRegistrationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainName := data.DomainName.ValueString()

	validateNameserversConfig(&resp.Diagnostics, data.Nameservers)
	if nameserversRemoved(data.Nameservers, state.Nameservers) {
		resp.Diagnostics.AddAttributeError(
			path.Root("nameservers"),
			"Cannot Clear Nameservers",
			"Route53 Domains does not support clearing nameservers through this provider. Set a replacement list of nameservers instead of removing the attribute.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	resourceTags, diags := frameworkMapToStringMap(ctx, data.Tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	desiredTags := mergeTags(r.defaultTags, resourceTags)
	addTagValidationDiagnostics(&resp.Diagnostics, path.Root("tags_all"), desiredTags)
	if resp.Diagnostics.HasError() {
		return
	}

	if tagManagementEnabled(r.defaultTags, data.Tags, state.Tags, state.TagsAll) {
		currentTags, err := r.listDomainTags(ctx, domainName)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error reading domain tags",
				fmt.Sprintf("Could not read tags for %s: %s", domainName, err.Error()),
			)
			return
		}

		err = r.syncDomainTags(ctx, domainName, currentTags, desiredTags)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating domain tags",
				fmt.Sprintf("Could not update tags for %s: %s", domainName, err.Error()),
			)
			return
		}
	}

	// Update auto-renew if changed
	if !data.AutoRenew.Equal(state.AutoRenew) {
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

	if !data.Nameservers.Equal(state.Nameservers) {
		nameservers, diags := frameworkListToAWSNameservers(ctx, data.Nameservers)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(nameservers) > 0 {
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
	}

	if !domainContactsEqual(data, state) {
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
	}

	if !domainPrivacySettingsEqual(data, state) {
		_, err := r.client.UpdateDomainContactPrivacy(ctx, &route53domains.UpdateDomainContactPrivacyInput{
			DomainName:        aws.String(domainName),
			AdminPrivacy:      aws.Bool(data.AdminPrivacy.ValueBool()),
			RegistrantPrivacy: aws.Bool(data.RegistrantPrivacy.ValueBool()),
			TechPrivacy:       aws.Bool(data.TechPrivacy.ValueBool()),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating privacy settings",
				fmt.Sprintf("Could not update privacy settings for %s: %s", domainName, err.Error()),
			)
			return
		}
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

	resp.Diagnostics.Append(r.populateDomainDetailState(ctx, &data, domainDetail)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.TagsAll, diags = stringMapToFrameworkMap(desiredTags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete removes the domain when allowed and performs best-effort zone cleanup.
func (r *DomainRegistrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DomainRegistrationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainName := data.DomainName.ValueString()

	// Check if deletion is allowed
	if !data.AllowDelete.ValueBool() {
		tflog.Warn(ctx, "Domain will be removed from state only (allow_delete = false)", map[string]any{
			"domain": domainName,
		})
		// Just remove from state, don't actually delete
		return
	}

	tflog.Warn(ctx, "DELETING DOMAIN REGISTRATION (allow_delete = true)", map[string]any{
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

	tflog.Info(ctx, "Domain deletion initiated", map[string]any{
		"domain": domainName,
	})

	// Attempt to delete the registrar-created hosted zone (safe - only deletes if all safeguards pass)
	err = r.deleteRegistrarHostedZone(ctx, domainName)
	if err != nil {
		tflog.Warn(ctx, "Could not delete hosted zone", map[string]any{
			"domain": domainName,
			"error":  err.Error(),
		})
		// Don't fail the destroy - domain is already deleted, zone cleanup is best-effort
	} else {
		tflog.Info(ctx, "Hosted zone deleted", map[string]any{
			"domain": domainName,
		})
	}
}

// ImportState imports the domain resource by its domain name.
func (r *DomainRegistrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("domain_name"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
