package provider

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	route53types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/aws/aws-sdk-go-v2/service/route53domains"
	"github.com/aws/aws-sdk-go-v2/service/route53domains/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	tftypes "github.com/hashicorp/terraform-plugin-framework/types"
)

// MockRoute53DomainsClient is a mock implementation for testing
type MockRoute53DomainsClient struct {
	GetDomainDetailFunc            func(ctx context.Context, params *route53domains.GetDomainDetailInput, optFns ...func(*route53domains.Options)) (*route53domains.GetDomainDetailOutput, error)
	RegisterDomainFunc             func(ctx context.Context, params *route53domains.RegisterDomainInput, optFns ...func(*route53domains.Options)) (*route53domains.RegisterDomainOutput, error)
	GetOperationDetailFunc         func(ctx context.Context, params *route53domains.GetOperationDetailInput, optFns ...func(*route53domains.Options)) (*route53domains.GetOperationDetailOutput, error)
	UpdateDomainNameserversFunc    func(ctx context.Context, params *route53domains.UpdateDomainNameserversInput, optFns ...func(*route53domains.Options)) (*route53domains.UpdateDomainNameserversOutput, error)
	EnableDomainAutoRenewFunc      func(ctx context.Context, params *route53domains.EnableDomainAutoRenewInput, optFns ...func(*route53domains.Options)) (*route53domains.EnableDomainAutoRenewOutput, error)
	DisableDomainAutoRenewFunc     func(ctx context.Context, params *route53domains.DisableDomainAutoRenewInput, optFns ...func(*route53domains.Options)) (*route53domains.DisableDomainAutoRenewOutput, error)
	UpdateDomainContactFunc        func(ctx context.Context, params *route53domains.UpdateDomainContactInput, optFns ...func(*route53domains.Options)) (*route53domains.UpdateDomainContactOutput, error)
	UpdateDomainContactPrivacyFunc func(ctx context.Context, params *route53domains.UpdateDomainContactPrivacyInput, optFns ...func(*route53domains.Options)) (*route53domains.UpdateDomainContactPrivacyOutput, error)
	DeleteDomainFunc               func(ctx context.Context, params *route53domains.DeleteDomainInput, optFns ...func(*route53domains.Options)) (*route53domains.DeleteDomainOutput, error)
	ListTagsForDomainFunc          func(ctx context.Context, params *route53domains.ListTagsForDomainInput, optFns ...func(*route53domains.Options)) (*route53domains.ListTagsForDomainOutput, error)
	UpdateTagsForDomainFunc        func(ctx context.Context, params *route53domains.UpdateTagsForDomainInput, optFns ...func(*route53domains.Options)) (*route53domains.UpdateTagsForDomainOutput, error)
	DeleteTagsForDomainFunc        func(ctx context.Context, params *route53domains.DeleteTagsForDomainInput, optFns ...func(*route53domains.Options)) (*route53domains.DeleteTagsForDomainOutput, error)
	CheckDomainAvailabilityFunc    func(ctx context.Context, params *route53domains.CheckDomainAvailabilityInput, optFns ...func(*route53domains.Options)) (*route53domains.CheckDomainAvailabilityOutput, error)
}

type MockRoute53Client struct {
	ListHostedZonesByNameFunc  func(ctx context.Context, params *route53.ListHostedZonesByNameInput, optFns ...func(*route53.Options)) (*route53.ListHostedZonesByNameOutput, error)
	ListResourceRecordSetsFunc func(ctx context.Context, params *route53.ListResourceRecordSetsInput, optFns ...func(*route53.Options)) (*route53.ListResourceRecordSetsOutput, error)
	DeleteHostedZoneFunc       func(ctx context.Context, params *route53.DeleteHostedZoneInput, optFns ...func(*route53.Options)) (*route53.DeleteHostedZoneOutput, error)
}

var (
	errListHostedZonesByNameFuncNotConfigured  = errors.New("ListHostedZonesByNameFunc not configured")
	errListResourceRecordSetsFuncNotConfigured = errors.New("ListResourceRecordSetsFunc not configured")
	errDeleteHostedZoneFuncNotConfigured       = errors.New("DeleteHostedZoneFunc not configured")
	errUnexpectedMockRoute53Call               = errors.New("unexpected Route53 mock call")
	errUnexpectedMockRoute53DomainsCall        = errors.New("unexpected Route53 Domains mock call")
	errMockAccessDenied                        = errors.New("access denied")
)

func (m *MockRoute53DomainsClient) GetDomainDetail(ctx context.Context, params *route53domains.GetDomainDetailInput, optFns ...func(*route53domains.Options)) (*route53domains.GetDomainDetailOutput, error) {
	if m.GetDomainDetailFunc == nil {
		return nil, errUnexpectedMockRoute53DomainsCall
	}
	return m.GetDomainDetailFunc(ctx, params, optFns...)
}

func (m *MockRoute53DomainsClient) RegisterDomain(ctx context.Context, params *route53domains.RegisterDomainInput, optFns ...func(*route53domains.Options)) (*route53domains.RegisterDomainOutput, error) {
	if m.RegisterDomainFunc == nil {
		return nil, errUnexpectedMockRoute53DomainsCall
	}
	return m.RegisterDomainFunc(ctx, params, optFns...)
}

func (m *MockRoute53DomainsClient) GetOperationDetail(ctx context.Context, params *route53domains.GetOperationDetailInput, optFns ...func(*route53domains.Options)) (*route53domains.GetOperationDetailOutput, error) {
	if m.GetOperationDetailFunc == nil {
		return nil, errUnexpectedMockRoute53DomainsCall
	}
	return m.GetOperationDetailFunc(ctx, params, optFns...)
}

func (m *MockRoute53DomainsClient) UpdateDomainNameservers(ctx context.Context, params *route53domains.UpdateDomainNameserversInput, optFns ...func(*route53domains.Options)) (*route53domains.UpdateDomainNameserversOutput, error) {
	if m.UpdateDomainNameserversFunc == nil {
		return nil, errUnexpectedMockRoute53DomainsCall
	}
	return m.UpdateDomainNameserversFunc(ctx, params, optFns...)
}

func (m *MockRoute53DomainsClient) EnableDomainAutoRenew(ctx context.Context, params *route53domains.EnableDomainAutoRenewInput, optFns ...func(*route53domains.Options)) (*route53domains.EnableDomainAutoRenewOutput, error) {
	if m.EnableDomainAutoRenewFunc == nil {
		return nil, errUnexpectedMockRoute53DomainsCall
	}
	return m.EnableDomainAutoRenewFunc(ctx, params, optFns...)
}

func (m *MockRoute53DomainsClient) DisableDomainAutoRenew(ctx context.Context, params *route53domains.DisableDomainAutoRenewInput, optFns ...func(*route53domains.Options)) (*route53domains.DisableDomainAutoRenewOutput, error) {
	if m.DisableDomainAutoRenewFunc == nil {
		return nil, errUnexpectedMockRoute53DomainsCall
	}
	return m.DisableDomainAutoRenewFunc(ctx, params, optFns...)
}

func (m *MockRoute53DomainsClient) UpdateDomainContact(ctx context.Context, params *route53domains.UpdateDomainContactInput, optFns ...func(*route53domains.Options)) (*route53domains.UpdateDomainContactOutput, error) {
	if m.UpdateDomainContactFunc == nil {
		return nil, errUnexpectedMockRoute53DomainsCall
	}
	return m.UpdateDomainContactFunc(ctx, params, optFns...)
}

func (m *MockRoute53DomainsClient) UpdateDomainContactPrivacy(ctx context.Context, params *route53domains.UpdateDomainContactPrivacyInput, optFns ...func(*route53domains.Options)) (*route53domains.UpdateDomainContactPrivacyOutput, error) {
	if m.UpdateDomainContactPrivacyFunc == nil {
		return nil, errUnexpectedMockRoute53DomainsCall
	}
	return m.UpdateDomainContactPrivacyFunc(ctx, params, optFns...)
}

func (m *MockRoute53DomainsClient) DeleteDomain(ctx context.Context, params *route53domains.DeleteDomainInput, optFns ...func(*route53domains.Options)) (*route53domains.DeleteDomainOutput, error) {
	if m.DeleteDomainFunc == nil {
		return nil, errUnexpectedMockRoute53DomainsCall
	}
	return m.DeleteDomainFunc(ctx, params, optFns...)
}

func (m *MockRoute53DomainsClient) ListTagsForDomain(ctx context.Context, params *route53domains.ListTagsForDomainInput, optFns ...func(*route53domains.Options)) (*route53domains.ListTagsForDomainOutput, error) {
	if m.ListTagsForDomainFunc == nil {
		return nil, errUnexpectedMockRoute53DomainsCall
	}
	return m.ListTagsForDomainFunc(ctx, params, optFns...)
}

func (m *MockRoute53DomainsClient) UpdateTagsForDomain(ctx context.Context, params *route53domains.UpdateTagsForDomainInput, optFns ...func(*route53domains.Options)) (*route53domains.UpdateTagsForDomainOutput, error) {
	if m.UpdateTagsForDomainFunc == nil {
		return nil, errUnexpectedMockRoute53DomainsCall
	}
	return m.UpdateTagsForDomainFunc(ctx, params, optFns...)
}

func (m *MockRoute53DomainsClient) DeleteTagsForDomain(ctx context.Context, params *route53domains.DeleteTagsForDomainInput, optFns ...func(*route53domains.Options)) (*route53domains.DeleteTagsForDomainOutput, error) {
	if m.DeleteTagsForDomainFunc == nil {
		return nil, errUnexpectedMockRoute53DomainsCall
	}
	return m.DeleteTagsForDomainFunc(ctx, params, optFns...)
}

func (m *MockRoute53Client) ListHostedZonesByName(ctx context.Context, params *route53.ListHostedZonesByNameInput, optFns ...func(*route53.Options)) (*route53.ListHostedZonesByNameOutput, error) {
	if m.ListHostedZonesByNameFunc == nil {
		return nil, errListHostedZonesByNameFuncNotConfigured
	}
	return m.ListHostedZonesByNameFunc(ctx, params, optFns...)
}

func (m *MockRoute53Client) ListResourceRecordSets(ctx context.Context, params *route53.ListResourceRecordSetsInput, optFns ...func(*route53.Options)) (*route53.ListResourceRecordSetsOutput, error) {
	if m.ListResourceRecordSetsFunc == nil {
		return nil, errListResourceRecordSetsFuncNotConfigured
	}
	return m.ListResourceRecordSetsFunc(ctx, params, optFns...)
}

func (m *MockRoute53Client) DeleteHostedZone(ctx context.Context, params *route53.DeleteHostedZoneInput, optFns ...func(*route53.Options)) (*route53.DeleteHostedZoneOutput, error) {
	if m.DeleteHostedZoneFunc == nil {
		return nil, errDeleteHostedZoneFuncNotConfigured
	}
	return m.DeleteHostedZoneFunc(ctx, params, optFns...)
}

func TestResourceSchema(t *testing.T) {
	ctx := context.Background()
	r := NewDomainRegistrationResource()

	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema returned errors: %v", resp.Diagnostics)
	}

	// Verify required attributes exist
	requiredAttrs := []string{
		"id",
		"domain_name",
		"duration_years",
		"auto_renew",
		"admin_contact",
		"registrant_contact",
		"tech_contact",
		"admin_privacy",
		"registrant_privacy",
		"tech_privacy",
		"nameservers",
		"tags",
		"tags_all",
		"allow_delete",
		"delete_hosted_zone",
		"status",
		"expiration_date",
		"creation_date",
		"registration_timeout",
		"registration_operation_id",
		"hosted_zone_id",
	}

	for _, attr := range requiredAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema missing '%s' attribute", attr)
		}
	}

	nameserversAttr, ok := resp.Schema.Attributes["nameservers"].(resourceschema.ListAttribute)
	if !ok {
		t.Fatalf("nameservers attribute has type %T, want schema.ListAttribute", resp.Schema.Attributes["nameservers"])
	}
	if !nameserversAttr.Optional {
		t.Fatal("nameservers attribute must be optional")
	}
	if !nameserversAttr.Computed {
		t.Fatal("nameservers attribute must be computed because Read stores AWS nameservers in state")
	}
}

func TestResourceModelNameserversAcceptsUnknownList(t *testing.T) {
	field, ok := reflect.TypeFor[DomainRegistrationResourceModel]().FieldByName("Nameservers")
	if !ok {
		t.Fatal("DomainRegistrationResourceModel missing Nameservers field")
	}
	if got, want := field.Type, reflect.TypeFor[tftypes.List](); got != want {
		t.Fatalf("Nameservers field type = %s, want %s", got, want)
	}
}

func TestResourceMetadata(t *testing.T) {
	ctx := context.Background()
	r := NewDomainRegistrationResource()

	req := resource.MetadataRequest{
		ProviderTypeName: "awsdomains",
	}
	resp := &resource.MetadataResponse{}
	r.Metadata(ctx, req, resp)

	expected := "awsdomains_domain"
	if resp.TypeName != expected {
		t.Errorf("Expected TypeName '%s', got '%s'", expected, resp.TypeName)
	}
}

func TestContactModelToAWS(t *testing.T) {
	tests := []struct {
		name     string
		input    *ContactModel
		expected *types.ContactDetail
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name: "full contact",
			input: &ContactModel{
				FirstName:    stringValue("John"),
				LastName:     stringValue("Doe"),
				Email:        stringValue("john@example.com"),
				PhoneNumber:  stringValue("+1.5551234567"),
				AddressLine1: stringValue("123 Main St"),
				City:         stringValue("Seattle"),
				State:        stringValue("WA"),
				ZipCode:      stringValue("98101"),
				CountryCode:  stringValue("US"),
				ContactType:  stringValue("PERSON"),
			},
			expected: &types.ContactDetail{
				FirstName:    aws.String("John"),
				LastName:     aws.String("Doe"),
				Email:        aws.String("john@example.com"),
				PhoneNumber:  aws.String("+1.5551234567"),
				AddressLine1: aws.String("123 Main St"),
				City:         aws.String("Seattle"),
				State:        aws.String("WA"),
				ZipCode:      aws.String("98101"),
				CountryCode:  types.CountryCodeUs,
				ContactType:  types.ContactTypePerson,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contactModelToAWS(tt.input)
			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
				return
			}
			if result == nil {
				t.Fatal("Expected non-nil result")
				return
			}
			if aws.ToString(result.FirstName) != aws.ToString(tt.expected.FirstName) {
				t.Errorf("FirstName mismatch: got %s, want %s", aws.ToString(result.FirstName), aws.ToString(tt.expected.FirstName))
			}
			if aws.ToString(result.LastName) != aws.ToString(tt.expected.LastName) {
				t.Errorf("LastName mismatch: got %s, want %s", aws.ToString(result.LastName), aws.ToString(tt.expected.LastName))
			}
			if aws.ToString(result.Email) != aws.ToString(tt.expected.Email) {
				t.Errorf("Email mismatch: got %s, want %s", aws.ToString(result.Email), aws.ToString(tt.expected.Email))
			}
		})
	}
}

func TestContactTypeDefault(t *testing.T) {
	// Test that empty contact type defaults to PERSON
	input := &ContactModel{
		FirstName:    stringValue("John"),
		LastName:     stringValue("Doe"),
		Email:        stringValue("john@example.com"),
		PhoneNumber:  stringValue("+1.5551234567"),
		AddressLine1: stringValue("123 Main St"),
		City:         stringValue("Seattle"),
		State:        stringValue("WA"),
		ZipCode:      stringValue("98101"),
		CountryCode:  stringValue("US"),
		// ContactType intentionally omitted
	}

	result := contactModelToAWS(input)
	if result.ContactType != types.ContactTypePerson {
		t.Errorf("Expected default ContactType 'PERSON', got '%s'", result.ContactType)
	}
}

func TestFrameworkListToAWSNameservers(t *testing.T) {
	ctx := context.Background()
	input, diags := tftypes.ListValueFrom(ctx, tftypes.StringType, []string{
		"ns1.example.com",
		"ns2.example.com",
	})
	if diags.HasError() {
		t.Fatalf("creating nameserver list returned diagnostics: %v", diags)
	}

	got, diags := frameworkListToAWSNameservers(ctx, input)
	if diags.HasError() {
		t.Fatalf("frameworkListToAWSNameservers returned diagnostics: %v", diags)
	}
	if len(got) != 2 {
		t.Fatalf("nameserver count = %d, want 2", len(got))
	}
	if aws.ToString(got[0].Name) != "ns1.example.com" {
		t.Fatalf("first nameserver = %q, want %q", aws.ToString(got[0].Name), "ns1.example.com")
	}
	if aws.ToString(got[1].Name) != "ns2.example.com" {
		t.Fatalf("second nameserver = %q, want %q", aws.ToString(got[1].Name), "ns2.example.com")
	}
}

func TestFrameworkListToAWSNameserversAllowsUnknownList(t *testing.T) {
	got, diags := frameworkListToAWSNameservers(context.Background(), tftypes.ListUnknown(tftypes.StringType))
	if diags.HasError() {
		t.Fatalf("frameworkListToAWSNameservers returned diagnostics: %v", diags)
	}
	if got != nil {
		t.Fatalf("nameservers = %#v, want nil", got)
	}
}

func TestDomainMutableSettingsUnchangedForTagOnlyUpdate(t *testing.T) {
	nameservers := stringListValue(t, "ns1.example.com", "ns2.example.com")
	state := DomainRegistrationResourceModel{
		AdminContact:      testContactModel("admin@example.com"),
		RegistrantContact: testContactModel("registrant@example.com"),
		TechContact:       testContactModel("tech@example.com"),
		AdminPrivacy:      tftypes.BoolValue(true),
		RegistrantPrivacy: tftypes.BoolValue(true),
		TechPrivacy:       tftypes.BoolValue(true),
		Nameservers:       nameservers,
	}
	plan := state

	if !domainContactsEqual(plan, state) {
		t.Fatal("domainContactsEqual returned false for tag-only update")
	}
	if !domainPrivacySettingsEqual(plan, state) {
		t.Fatal("domainPrivacySettingsEqual returned false for tag-only update")
	}
	if !plan.Nameservers.Equal(state.Nameservers) {
		t.Fatal("nameservers comparison returned false for tag-only update")
	}
}

func TestDomainContactsEqualDetectsContactChange(t *testing.T) {
	state := DomainRegistrationResourceModel{
		AdminContact:      testContactModel("admin@example.com"),
		RegistrantContact: testContactModel("registrant@example.com"),
		TechContact:       testContactModel("tech@example.com"),
	}
	plan := DomainRegistrationResourceModel{
		AdminContact:      testContactModel("new-admin@example.com"),
		RegistrantContact: testContactModel("registrant@example.com"),
		TechContact:       testContactModel("tech@example.com"),
	}

	if domainContactsEqual(plan, state) {
		t.Fatal("domainContactsEqual returned true for changed contact")
	}
}

func TestDomainPrivacySettingsEqualDetectsPrivacyChange(t *testing.T) {
	state := DomainRegistrationResourceModel{
		AdminPrivacy:      tftypes.BoolValue(true),
		RegistrantPrivacy: tftypes.BoolValue(true),
		TechPrivacy:       tftypes.BoolValue(true),
	}
	plan := DomainRegistrationResourceModel{
		AdminPrivacy:      tftypes.BoolValue(true),
		RegistrantPrivacy: tftypes.BoolValue(false),
		TechPrivacy:       tftypes.BoolValue(true),
	}

	if domainPrivacySettingsEqual(plan, state) {
		t.Fatal("domainPrivacySettingsEqual returned true for changed privacy setting")
	}
}

func TestUpdateTagOnlySkipsUnchangedDomainMutations(t *testing.T) {
	ctx := context.Background()
	state := testDomainModel(t, "tagless.example.com")
	state.Tags = stringMapValue(t, map[string]string{"Environment": "dev"})
	state.TagsAll = stringMapValue(t, map[string]string{"Environment": "dev"})

	plan := state
	plan.Tags = stringMapValue(t, map[string]string{"Environment": "prod"})
	plan.TagsAll = stringMapValue(t, map[string]string{"Environment": "prod"})

	updateTagsCalled := false
	domainResource := &DomainRegistrationResource{
		client: &MockRoute53DomainsClient{
			ListTagsForDomainFunc: func(_ context.Context, _ *route53domains.ListTagsForDomainInput, _ ...func(*route53domains.Options)) (*route53domains.ListTagsForDomainOutput, error) {
				return &route53domains.ListTagsForDomainOutput{
					TagList: []types.Tag{{Key: aws.String("Environment"), Value: aws.String("dev")}},
				}, nil
			},
			UpdateTagsForDomainFunc: func(_ context.Context, params *route53domains.UpdateTagsForDomainInput, _ ...func(*route53domains.Options)) (*route53domains.UpdateTagsForDomainOutput, error) {
				updateTagsCalled = true
				if len(params.TagsToUpdate) != 1 {
					t.Fatalf("TagsToUpdate count = %d, want 1", len(params.TagsToUpdate))
				}
				if got := aws.ToString(params.TagsToUpdate[0].Value); got != "prod" {
					t.Fatalf("updated tag value = %q, want %q", got, "prod")
				}
				return &route53domains.UpdateTagsForDomainOutput{}, nil
			},
			UpdateDomainNameserversFunc: func(context.Context, *route53domains.UpdateDomainNameserversInput, ...func(*route53domains.Options)) (*route53domains.UpdateDomainNameserversOutput, error) {
				t.Fatal("UpdateDomainNameservers must not be called for tag-only update")
				return nil, errUnexpectedMockRoute53DomainsCall
			},
			UpdateDomainContactFunc: func(context.Context, *route53domains.UpdateDomainContactInput, ...func(*route53domains.Options)) (*route53domains.UpdateDomainContactOutput, error) {
				t.Fatal("UpdateDomainContact must not be called for tag-only update")
				return nil, errUnexpectedMockRoute53DomainsCall
			},
			UpdateDomainContactPrivacyFunc: func(context.Context, *route53domains.UpdateDomainContactPrivacyInput, ...func(*route53domains.Options)) (*route53domains.UpdateDomainContactPrivacyOutput, error) {
				t.Fatal("UpdateDomainContactPrivacy must not be called for tag-only update")
				return nil, errUnexpectedMockRoute53DomainsCall
			},
			GetDomainDetailFunc: func(_ context.Context, _ *route53domains.GetDomainDetailInput, _ ...func(*route53domains.Options)) (*route53domains.GetDomainDetailOutput, error) {
				return MockDomainDetailResponse("tagless.example.com"), nil
			},
		},
	}

	schema := testDomainResourceSchema(t)
	req := resourceUpdateRequest(t, schema, plan, state)
	resp := &resource.UpdateResponse{State: tfsdk.State{Schema: schema}}

	domainResource.Update(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Update returned diagnostics: %v", resp.Diagnostics)
	}
	if !updateTagsCalled {
		t.Fatal("UpdateTagsForDomain was not called")
	}
}

func TestUpdateRejectsNameserverRemoval(t *testing.T) {
	state := testDomainModel(t, "example.com")
	plan := state
	plan.Nameservers = tftypes.ListNull(tftypes.StringType)

	schema := testDomainResourceSchema(t)
	req := resourceUpdateRequest(t, schema, plan, state)
	resp := &resource.UpdateResponse{State: tfsdk.State{Schema: schema}}

	domainResource := &DomainRegistrationResource{client: &MockRoute53DomainsClient{}}
	domainResource.Update(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected nameserver removal to return an error")
	}
}

func TestReadSkipsTagAPIWhenTagsUnmanaged(t *testing.T) {
	ctx := context.Background()
	state := testDomainModel(t, "example.com")
	state.Tags = emptyFrameworkStringMap()
	state.TagsAll = emptyFrameworkStringMap()

	domainResource := &DomainRegistrationResource{
		client: &MockRoute53DomainsClient{
			GetDomainDetailFunc: func(_ context.Context, _ *route53domains.GetDomainDetailInput, _ ...func(*route53domains.Options)) (*route53domains.GetDomainDetailOutput, error) {
				return MockDomainDetailResponse("example.com"), nil
			},
			ListTagsForDomainFunc: func(context.Context, *route53domains.ListTagsForDomainInput, ...func(*route53domains.Options)) (*route53domains.ListTagsForDomainOutput, error) {
				t.Fatal("ListTagsForDomain must not be called when tags are unmanaged")
				return nil, errUnexpectedMockRoute53DomainsCall
			},
		},
		route53Client: &MockRoute53Client{
			ListHostedZonesByNameFunc: func(context.Context, *route53.ListHostedZonesByNameInput, ...func(*route53.Options)) (*route53.ListHostedZonesByNameOutput, error) {
				return &route53.ListHostedZonesByNameOutput{}, nil
			},
		},
	}

	schema := testDomainResourceSchema(t)
	req := resourceReadRequest(t, schema, state)
	resp := &resource.ReadResponse{State: tfsdk.State{Schema: schema}}

	domainResource.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Read returned diagnostics: %v", resp.Diagnostics)
	}
}

func TestReadHydratesContactAndPrivacyDrift(t *testing.T) {
	ctx := context.Background()
	state := testDomainModel(t, "example.com")
	state.AdminContact.Email = stringValue("old-admin@example.com")

	domainResource := &DomainRegistrationResource{
		client: &MockRoute53DomainsClient{
			GetDomainDetailFunc: func(_ context.Context, _ *route53domains.GetDomainDetailInput, _ ...func(*route53domains.Options)) (*route53domains.GetDomainDetailOutput, error) {
				detail := MockDomainDetailResponse("example.com")
				detail.AdminContact.Email = aws.String("new-admin@example.com")
				detail.AdminPrivacy = aws.Bool(false)
				return detail, nil
			},
		},
		route53Client: &MockRoute53Client{
			ListHostedZonesByNameFunc: func(context.Context, *route53.ListHostedZonesByNameInput, ...func(*route53.Options)) (*route53.ListHostedZonesByNameOutput, error) {
				return &route53.ListHostedZonesByNameOutput{}, nil
			},
		},
	}

	schema := testDomainResourceSchema(t)
	req := resourceReadRequest(t, schema, state)
	resp := &resource.ReadResponse{State: tfsdk.State{Schema: schema}}

	domainResource.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Read returned diagnostics: %v", resp.Diagnostics)
	}

	var got DomainRegistrationResourceModel
	resp.Diagnostics.Append(resp.State.Get(ctx, &got)...)
	if resp.Diagnostics.HasError() {
		t.Fatalf("reading response state returned diagnostics: %v", resp.Diagnostics)
	}
	if got.AdminContact.Email.ValueString() != "new-admin@example.com" {
		t.Fatalf("admin email = %q, want %q", got.AdminContact.Email.ValueString(), "new-admin@example.com")
	}
	if got.AdminPrivacy.ValueBool() {
		t.Fatal("admin privacy was not refreshed from AWS")
	}
}

func TestCreateWarnsAndKeepsStateWhenPostRegistrationTagSyncFails(t *testing.T) {
	ctx := context.Background()
	plan := testDomainModel(t, "example.com")
	plan.Tags = stringMapValue(t, map[string]string{"Environment": "prod"})
	plan.TagsAll = stringMapValue(t, map[string]string{"Environment": "prod"})
	plan.Nameservers = tftypes.ListNull(tftypes.StringType)

	domainResource := &DomainRegistrationResource{
		client: &MockRoute53DomainsClient{
			RegisterDomainFunc: func(_ context.Context, _ *route53domains.RegisterDomainInput, _ ...func(*route53domains.Options)) (*route53domains.RegisterDomainOutput, error) {
				return &route53domains.RegisterDomainOutput{OperationId: aws.String("op-123")}, nil
			},
			GetOperationDetailFunc: func(_ context.Context, _ *route53domains.GetOperationDetailInput, _ ...func(*route53domains.Options)) (*route53domains.GetOperationDetailOutput, error) {
				return &route53domains.GetOperationDetailOutput{Status: types.OperationStatusSuccessful}, nil
			},
			UpdateTagsForDomainFunc: func(context.Context, *route53domains.UpdateTagsForDomainInput, ...func(*route53domains.Options)) (*route53domains.UpdateTagsForDomainOutput, error) {
				return nil, errMockAccessDenied
			},
			ListTagsForDomainFunc: func(context.Context, *route53domains.ListTagsForDomainInput, ...func(*route53domains.Options)) (*route53domains.ListTagsForDomainOutput, error) {
				t.Fatal("ListTagsForDomain must not be called after create tag sync failure; state must keep planned tags_all")
				return nil, errUnexpectedMockRoute53DomainsCall
			},
			GetDomainDetailFunc: func(_ context.Context, _ *route53domains.GetDomainDetailInput, _ ...func(*route53domains.Options)) (*route53domains.GetDomainDetailOutput, error) {
				return MockDomainDetailResponse("example.com"), nil
			},
		},
		route53Client: &MockRoute53Client{
			ListHostedZonesByNameFunc: func(context.Context, *route53.ListHostedZonesByNameInput, ...func(*route53.Options)) (*route53.ListHostedZonesByNameOutput, error) {
				return &route53.ListHostedZonesByNameOutput{}, nil
			},
		},
	}

	schema := testDomainResourceSchema(t)
	req := resourceCreateRequest(t, schema, plan)
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: schema}}

	domainResource.Create(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error diagnostics: %v", resp.Diagnostics)
	}
	if len(resp.Diagnostics) == 0 {
		t.Fatal("expected warning diagnostic for failed post-registration tag sync")
	}

	var got DomainRegistrationResourceModel
	resp.Diagnostics.Append(resp.State.Get(ctx, &got)...)
	if resp.Diagnostics.HasError() {
		t.Fatalf("reading response state returned diagnostics: %v", resp.Diagnostics)
	}
	if got.ID.ValueString() != "example.com" {
		t.Fatalf("id = %q, want %q", got.ID.ValueString(), "example.com")
	}
	tagsAll, diags := frameworkMapToStringMap(ctx, got.TagsAll)
	if diags.HasError() {
		t.Fatalf("reading tags_all returned diagnostics: %v", diags)
	}
	if tagsAll["Environment"] != "prod" {
		t.Fatalf("tags_all[Environment] = %q, want %q", tagsAll["Environment"], "prod")
	}
}

func TestCreateKeepsStateWhenRegistrationStatusUnknown(t *testing.T) {
	ctx := context.Background()
	plan := testDomainModel(t, "example.com")
	plan.RegistrationTimeout = tftypes.Int64Value(0)
	plan.Nameservers = tftypes.ListNull(tftypes.StringType)

	domainResource := &DomainRegistrationResource{
		client: &MockRoute53DomainsClient{
			RegisterDomainFunc: func(_ context.Context, _ *route53domains.RegisterDomainInput, _ ...func(*route53domains.Options)) (*route53domains.RegisterDomainOutput, error) {
				return &route53domains.RegisterDomainOutput{OperationId: aws.String("op-123")}, nil
			},
			GetOperationDetailFunc: func(_ context.Context, _ *route53domains.GetOperationDetailInput, _ ...func(*route53domains.Options)) (*route53domains.GetOperationDetailOutput, error) {
				return &route53domains.GetOperationDetailOutput{Status: types.OperationStatusInProgress}, nil
			},
			UpdateTagsForDomainFunc: func(context.Context, *route53domains.UpdateTagsForDomainInput, ...func(*route53domains.Options)) (*route53domains.UpdateTagsForDomainOutput, error) {
				t.Fatal("UpdateTagsForDomain must not be called until registration completion is confirmed")
				return nil, errUnexpectedMockRoute53DomainsCall
			},
			UpdateDomainNameserversFunc: func(context.Context, *route53domains.UpdateDomainNameserversInput, ...func(*route53domains.Options)) (*route53domains.UpdateDomainNameserversOutput, error) {
				t.Fatal("UpdateDomainNameservers must not be called until registration completion is confirmed")
				return nil, errUnexpectedMockRoute53DomainsCall
			},
			GetDomainDetailFunc: func(context.Context, *route53domains.GetDomainDetailInput, ...func(*route53domains.Options)) (*route53domains.GetDomainDetailOutput, error) {
				t.Fatal("GetDomainDetail must not be called until registration completion is confirmed")
				return nil, errUnexpectedMockRoute53DomainsCall
			},
		},
	}

	schema := testDomainResourceSchema(t)
	req := resourceCreateRequest(t, schema, plan)
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: schema}}

	domainResource.Create(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error diagnostics: %v", resp.Diagnostics)
	}
	if len(resp.Diagnostics) == 0 {
		t.Fatal("expected warning diagnostic for unknown registration status")
	}

	var got DomainRegistrationResourceModel
	resp.Diagnostics.Append(resp.State.Get(ctx, &got)...)
	if resp.Diagnostics.HasError() {
		t.Fatalf("reading response state returned diagnostics: %v", resp.Diagnostics)
	}
	if got.ID.ValueString() != "example.com" {
		t.Fatalf("id = %q, want %q", got.ID.ValueString(), "example.com")
	}
	if got.Status.ValueString() != string(types.OperationStatusInProgress) {
		t.Fatalf("status = %q, want %q", got.Status.ValueString(), types.OperationStatusInProgress)
	}
	if got.RegistrationOperationID.ValueString() != "op-123" {
		t.Fatalf("registration_operation_id = %q, want %q", got.RegistrationOperationID.ValueString(), "op-123")
	}
}

func TestCreateWarnsAndKeepsStateWhenDomainDetailRefreshFails(t *testing.T) {
	ctx := context.Background()
	plan := testDomainModel(t, "example.com")
	plan.Nameservers = tftypes.ListNull(tftypes.StringType)

	domainResource := &DomainRegistrationResource{
		client: &MockRoute53DomainsClient{
			RegisterDomainFunc: func(_ context.Context, _ *route53domains.RegisterDomainInput, _ ...func(*route53domains.Options)) (*route53domains.RegisterDomainOutput, error) {
				return &route53domains.RegisterDomainOutput{OperationId: aws.String("op-123")}, nil
			},
			GetOperationDetailFunc: func(_ context.Context, _ *route53domains.GetOperationDetailInput, _ ...func(*route53domains.Options)) (*route53domains.GetOperationDetailOutput, error) {
				return &route53domains.GetOperationDetailOutput{Status: types.OperationStatusSuccessful}, nil
			},
			GetDomainDetailFunc: func(context.Context, *route53domains.GetDomainDetailInput, ...func(*route53domains.Options)) (*route53domains.GetDomainDetailOutput, error) {
				return nil, errMockAccessDenied
			},
		},
		route53Client: &MockRoute53Client{
			ListHostedZonesByNameFunc: func(context.Context, *route53.ListHostedZonesByNameInput, ...func(*route53.Options)) (*route53.ListHostedZonesByNameOutput, error) {
				return &route53.ListHostedZonesByNameOutput{}, nil
			},
		},
	}

	schema := testDomainResourceSchema(t)
	req := resourceCreateRequest(t, schema, plan)
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: schema}}

	domainResource.Create(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Create returned error diagnostics: %v", resp.Diagnostics)
	}
	if len(resp.Diagnostics) == 0 {
		t.Fatal("expected warning diagnostic for failed detail refresh")
	}

	var got DomainRegistrationResourceModel
	resp.Diagnostics.Append(resp.State.Get(ctx, &got)...)
	if resp.Diagnostics.HasError() {
		t.Fatalf("reading response state returned diagnostics: %v", resp.Diagnostics)
	}
	if got.ID.ValueString() != "example.com" {
		t.Fatalf("id = %q, want %q", got.ID.ValueString(), "example.com")
	}
	if got.Status.ValueString() != string(types.OperationStatusSuccessful) {
		t.Fatalf("status = %q, want %q", got.Status.ValueString(), types.OperationStatusSuccessful)
	}
	if got.RegistrationOperationID.ValueString() != "op-123" {
		t.Fatalf("registration_operation_id = %q, want %q", got.RegistrationOperationID.ValueString(), "op-123")
	}
}

func TestReadKeepsPendingRegistrationStateWhenDomainDetailFails(t *testing.T) {
	ctx := context.Background()
	state := testDomainModel(t, "example.com")
	state.Status = tftypes.StringValue(string(types.OperationStatusInProgress))
	state.RegistrationOperationID = tftypes.StringValue("op-123")

	domainResource := &DomainRegistrationResource{
		client: &MockRoute53DomainsClient{
			GetDomainDetailFunc: func(context.Context, *route53domains.GetDomainDetailInput, ...func(*route53domains.Options)) (*route53domains.GetDomainDetailOutput, error) {
				return nil, errMockAccessDenied
			},
			GetOperationDetailFunc: func(_ context.Context, params *route53domains.GetOperationDetailInput, _ ...func(*route53domains.Options)) (*route53domains.GetOperationDetailOutput, error) {
				if got := aws.ToString(params.OperationId); got != "op-123" {
					t.Fatalf("operation ID = %q, want %q", got, "op-123")
				}
				return &route53domains.GetOperationDetailOutput{Status: types.OperationStatusInProgress}, nil
			},
		},
	}

	schema := testDomainResourceSchema(t)
	req := resourceReadRequest(t, schema, state)
	resp := &resource.ReadResponse{State: tfsdk.State{Schema: schema}}

	domainResource.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Read returned error diagnostics: %v", resp.Diagnostics)
	}
	if len(resp.Diagnostics) == 0 {
		t.Fatal("expected warning diagnostic for pending registration detail read failure")
	}

	var got DomainRegistrationResourceModel
	resp.Diagnostics.Append(resp.State.Get(ctx, &got)...)
	if resp.Diagnostics.HasError() {
		t.Fatalf("reading response state returned diagnostics: %v", resp.Diagnostics)
	}
	if got.ID.ValueString() != "example.com" {
		t.Fatalf("id = %q, want %q", got.ID.ValueString(), "example.com")
	}
	if got.Status.ValueString() != string(types.OperationStatusInProgress) {
		t.Fatalf("status = %q, want %q", got.Status.ValueString(), types.OperationStatusInProgress)
	}
}

func TestReadReturnsErrorForSuccessfulRegistrationWhenDomainDetailFails(t *testing.T) {
	ctx := context.Background()
	state := testDomainModel(t, "example.com")
	state.Status = tftypes.StringNull()
	state.RegistrationOperationID = tftypes.StringValue("op-123")

	domainResource := &DomainRegistrationResource{
		client: &MockRoute53DomainsClient{
			GetDomainDetailFunc: func(context.Context, *route53domains.GetDomainDetailInput, ...func(*route53domains.Options)) (*route53domains.GetDomainDetailOutput, error) {
				return nil, errMockAccessDenied
			},
			GetOperationDetailFunc: func(_ context.Context, params *route53domains.GetOperationDetailInput, _ ...func(*route53domains.Options)) (*route53domains.GetOperationDetailOutput, error) {
				if got := aws.ToString(params.OperationId); got != "op-123" {
					t.Fatalf("operation ID = %q, want %q", got, "op-123")
				}
				return &route53domains.GetOperationDetailOutput{Status: types.OperationStatusSuccessful}, nil
			},
		},
	}

	schema := testDomainResourceSchema(t)
	req := resourceReadRequest(t, schema, state)
	resp := &resource.ReadResponse{State: tfsdk.State{Schema: schema}}

	domainResource.Read(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected detail read failure to return an error after registration succeeded")
	}

	var got DomainRegistrationResourceModel
	diags := resp.State.Get(ctx, &got)
	if diags.HasError() {
		t.Fatalf("reading response state returned diagnostics: %v", diags)
	}
	if got.ID.ValueString() != "example.com" {
		t.Fatalf("id = %q, want %q", got.ID.ValueString(), "example.com")
	}
	if got.Status.ValueString() != string(types.OperationStatusSuccessful) {
		t.Fatalf("status = %q, want %q", got.Status.ValueString(), types.OperationStatusSuccessful)
	}
	if got.RegistrationOperationID.ValueString() != "op-123" {
		t.Fatalf("registration_operation_id = %q, want %q", got.RegistrationOperationID.ValueString(), "op-123")
	}
}

func TestReadClearsRegistrationOperationIDWhenDomainHydrates(t *testing.T) {
	ctx := context.Background()
	state := testDomainModel(t, "example.com")
	state.RegistrationOperationID = tftypes.StringValue("op-123")

	domainResource := &DomainRegistrationResource{
		client: &MockRoute53DomainsClient{
			GetDomainDetailFunc: func(context.Context, *route53domains.GetDomainDetailInput, ...func(*route53domains.Options)) (*route53domains.GetDomainDetailOutput, error) {
				return MockDomainDetailResponse("example.com"), nil
			},
		},
		route53Client: &MockRoute53Client{
			ListHostedZonesByNameFunc: func(context.Context, *route53.ListHostedZonesByNameInput, ...func(*route53.Options)) (*route53.ListHostedZonesByNameOutput, error) {
				return &route53.ListHostedZonesByNameOutput{}, nil
			},
		},
	}

	schema := testDomainResourceSchema(t)
	req := resourceReadRequest(t, schema, state)
	resp := &resource.ReadResponse{State: tfsdk.State{Schema: schema}}

	domainResource.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Read returned diagnostics: %v", resp.Diagnostics)
	}

	var got DomainRegistrationResourceModel
	resp.Diagnostics.Append(resp.State.Get(ctx, &got)...)
	if resp.Diagnostics.HasError() {
		t.Fatalf("reading response state returned diagnostics: %v", resp.Diagnostics)
	}
	if !got.RegistrationOperationID.IsNull() {
		t.Fatalf("registration_operation_id = %q, want null", got.RegistrationOperationID.ValueString())
	}
}

func TestReadRemovesPendingRegistrationStateWhenOperationFails(t *testing.T) {
	ctx := context.Background()
	state := testDomainModel(t, "example.com")
	state.Status = tftypes.StringValue(string(types.OperationStatusInProgress))
	state.RegistrationOperationID = tftypes.StringValue("op-123")

	domainResource := &DomainRegistrationResource{
		client: &MockRoute53DomainsClient{
			GetDomainDetailFunc: func(context.Context, *route53domains.GetDomainDetailInput, ...func(*route53domains.Options)) (*route53domains.GetDomainDetailOutput, error) {
				return nil, errMockAccessDenied
			},
			GetOperationDetailFunc: func(_ context.Context, params *route53domains.GetOperationDetailInput, _ ...func(*route53domains.Options)) (*route53domains.GetOperationDetailOutput, error) {
				if got := aws.ToString(params.OperationId); got != "op-123" {
					t.Fatalf("operation ID = %q, want %q", got, "op-123")
				}
				return &route53domains.GetOperationDetailOutput{
					Status:  types.OperationStatusFailed,
					Message: aws.String("registration rejected"),
				}, nil
			},
		},
	}

	schema := testDomainResourceSchema(t)
	req := resourceReadRequest(t, schema, state)
	resp := &resource.ReadResponse{State: tfsdk.State{Schema: schema}}

	domainResource.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Read returned error diagnostics: %v", resp.Diagnostics)
	}
	if len(resp.Diagnostics) == 0 {
		t.Fatal("expected warning diagnostic for failed registration operation")
	}
	if !resp.State.Raw.IsNull() {
		t.Fatalf("state was not removed after failed registration operation: %s", resp.State.Raw.String())
	}
}

func TestReadRemovesStateForDomainNotFound(t *testing.T) {
	ctx := context.Background()
	state := testDomainModel(t, "example.com")

	domainResource := &DomainRegistrationResource{
		client: &MockRoute53DomainsClient{
			GetDomainDetailFunc: func(context.Context, *route53domains.GetDomainDetailInput, ...func(*route53domains.Options)) (*route53domains.GetDomainDetailOutput, error) {
				return nil, &types.InvalidInput{Message: aws.String("Domain not found")}
			},
		},
	}

	schema := testDomainResourceSchema(t)
	req := resourceReadRequest(t, schema, state)
	resp := &resource.ReadResponse{State: tfsdk.State{Schema: schema}}

	domainResource.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Read returned error diagnostics: %v", resp.Diagnostics)
	}
	if !resp.State.Raw.IsNull() {
		t.Fatalf("state was not removed after domain not found: %s", resp.State.Raw.String())
	}
}

func TestUpdateWaitsForNameserverOperationBeforeRefreshingState(t *testing.T) {
	ctx := context.Background()
	state := testDomainModel(t, "example.com")
	plan := state
	plan.Nameservers = stringListValue(t, "ns3.example.com", "ns4.example.com")

	waitedForOperation := false
	domainResource := &DomainRegistrationResource{
		client: &MockRoute53DomainsClient{
			UpdateDomainNameserversFunc: func(_ context.Context, _ *route53domains.UpdateDomainNameserversInput, _ ...func(*route53domains.Options)) (*route53domains.UpdateDomainNameserversOutput, error) {
				return &route53domains.UpdateDomainNameserversOutput{OperationId: aws.String("op-ns")}, nil
			},
			GetOperationDetailFunc: func(_ context.Context, params *route53domains.GetOperationDetailInput, _ ...func(*route53domains.Options)) (*route53domains.GetOperationDetailOutput, error) {
				if got := aws.ToString(params.OperationId); got != "op-ns" {
					t.Fatalf("operation ID = %q, want %q", got, "op-ns")
				}
				waitedForOperation = true
				return &route53domains.GetOperationDetailOutput{Status: types.OperationStatusSuccessful}, nil
			},
			GetDomainDetailFunc: func(_ context.Context, _ *route53domains.GetDomainDetailInput, _ ...func(*route53domains.Options)) (*route53domains.GetDomainDetailOutput, error) {
				if !waitedForOperation {
					t.Fatal("GetDomainDetail was called before the nameserver operation completed")
				}
				detail := MockDomainDetailResponse("example.com")
				detail.Nameservers = []types.Nameserver{
					{Name: aws.String("ns3.example.com")},
					{Name: aws.String("ns4.example.com")},
				}
				return detail, nil
			},
		},
		route53Client: &MockRoute53Client{
			ListHostedZonesByNameFunc: func(context.Context, *route53.ListHostedZonesByNameInput, ...func(*route53.Options)) (*route53.ListHostedZonesByNameOutput, error) {
				return &route53.ListHostedZonesByNameOutput{}, nil
			},
		},
	}

	schema := testDomainResourceSchema(t)
	req := resourceUpdateRequest(t, schema, plan, state)
	resp := &resource.UpdateResponse{State: tfsdk.State{Schema: schema}}

	domainResource.Update(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Update returned diagnostics: %v", resp.Diagnostics)
	}
	if !waitedForOperation {
		t.Fatal("GetOperationDetail was not called")
	}
}

func TestUpdateReturnsErrorWhenNameserverOperationFails(t *testing.T) {
	ctx := context.Background()
	state := testDomainModel(t, "example.com")
	plan := state
	plan.Nameservers = stringListValue(t, "ns3.example.com", "ns4.example.com")

	domainResource := &DomainRegistrationResource{
		client: &MockRoute53DomainsClient{
			UpdateDomainNameserversFunc: func(_ context.Context, _ *route53domains.UpdateDomainNameserversInput, _ ...func(*route53domains.Options)) (*route53domains.UpdateDomainNameserversOutput, error) {
				return &route53domains.UpdateDomainNameserversOutput{OperationId: aws.String("op-ns")}, nil
			},
			GetOperationDetailFunc: func(_ context.Context, _ *route53domains.GetOperationDetailInput, _ ...func(*route53domains.Options)) (*route53domains.GetOperationDetailOutput, error) {
				return &route53domains.GetOperationDetailOutput{
					Status:  types.OperationStatusFailed,
					Message: aws.String("pending customer action"),
				}, nil
			},
			GetDomainDetailFunc: func(context.Context, *route53domains.GetDomainDetailInput, ...func(*route53domains.Options)) (*route53domains.GetDomainDetailOutput, error) {
				t.Fatal("GetDomainDetail must not be called after operation failure")
				return nil, errUnexpectedMockRoute53DomainsCall
			},
		},
	}

	schema := testDomainResourceSchema(t)
	req := resourceUpdateRequest(t, schema, plan, state)
	resp := &resource.UpdateResponse{State: tfsdk.State{Schema: schema}}

	domainResource.Update(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected operation failure diagnostic")
	}
}

func TestUpdateDeletesHostedZoneWhenDeleteHostedZoneEnabled(t *testing.T) {
	ctx := context.Background()
	state := testDomainModel(t, "example.com")
	state.DeleteHostedZone = tftypes.BoolValue(false)
	state.HostedZoneID = tftypes.StringValue("ZREG")
	plan := state
	plan.DeleteHostedZone = tftypes.BoolValue(true)

	deleteCallCount := 0
	domainResource := &DomainRegistrationResource{
		client: &MockRoute53DomainsClient{
			GetDomainDetailFunc: func(_ context.Context, _ *route53domains.GetDomainDetailInput, _ ...func(*route53domains.Options)) (*route53domains.GetDomainDetailOutput, error) {
				return MockDomainDetailResponse("example.com"), nil
			},
		},
		route53Client: &MockRoute53Client{
			ListHostedZonesByNameFunc: func(context.Context, *route53.ListHostedZonesByNameInput, ...func(*route53.Options)) (*route53.ListHostedZonesByNameOutput, error) {
				return &route53.ListHostedZonesByNameOutput{
					HostedZones: []route53types.HostedZone{
						{
							Id:              aws.String("/hostedzone/ZREG"),
							Name:            aws.String("example.com."),
							CallerReference: aws.String("RISWorkflow-RD:test"),
							Config:          &route53types.HostedZoneConfig{Comment: aws.String(registrarHostedZoneComment)},
						},
					},
				}, nil
			},
			ListResourceRecordSetsFunc: func(context.Context, *route53.ListResourceRecordSetsInput, ...func(*route53.Options)) (*route53.ListResourceRecordSetsOutput, error) {
				return &route53.ListResourceRecordSetsOutput{
					ResourceRecordSets: []route53types.ResourceRecordSet{
						{Name: aws.String("example.com."), Type: route53types.RRTypeNs},
						{Name: aws.String("example.com."), Type: route53types.RRTypeSoa},
					},
				}, nil
			},
			DeleteHostedZoneFunc: func(_ context.Context, params *route53.DeleteHostedZoneInput, _ ...func(*route53.Options)) (*route53.DeleteHostedZoneOutput, error) {
				if got := aws.ToString(params.Id); got != "/hostedzone/ZREG" {
					t.Fatalf("hosted zone ID = %q, want %q", got, "/hostedzone/ZREG")
				}
				deleteCallCount++
				return &route53.DeleteHostedZoneOutput{}, nil
			},
		},
	}

	schema := testDomainResourceSchema(t)
	req := resourceUpdateRequest(t, schema, plan, state)
	resp := &resource.UpdateResponse{State: tfsdk.State{Schema: schema}}

	domainResource.Update(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Update returned diagnostics: %v", resp.Diagnostics)
	}
	if deleteCallCount != 1 {
		t.Fatalf("DeleteHostedZone call count = %d, want %d", deleteCallCount, 1)
	}

	var got DomainRegistrationResourceModel
	resp.Diagnostics.Append(resp.State.Get(ctx, &got)...)
	if resp.Diagnostics.HasError() {
		t.Fatalf("reading response state returned diagnostics: %v", resp.Diagnostics)
	}
	if !got.HostedZoneID.IsNull() {
		t.Fatalf("hosted_zone_id = %q, want null", got.HostedZoneID.ValueString())
	}
}

func TestUpdateTreatsMissingHostedZoneAsDeleted(t *testing.T) {
	ctx := context.Background()
	state := testDomainModel(t, "example.com")
	state.DeleteHostedZone = tftypes.BoolValue(false)
	state.HostedZoneID = tftypes.StringValue("ZREG")
	plan := state
	plan.DeleteHostedZone = tftypes.BoolValue(true)

	domainResource := &DomainRegistrationResource{
		client: &MockRoute53DomainsClient{
			GetDomainDetailFunc: func(_ context.Context, _ *route53domains.GetDomainDetailInput, _ ...func(*route53domains.Options)) (*route53domains.GetDomainDetailOutput, error) {
				return MockDomainDetailResponse("example.com"), nil
			},
		},
		route53Client: &MockRoute53Client{
			ListHostedZonesByNameFunc: func(context.Context, *route53.ListHostedZonesByNameInput, ...func(*route53.Options)) (*route53.ListHostedZonesByNameOutput, error) {
				return &route53.ListHostedZonesByNameOutput{}, nil
			},
			ListResourceRecordSetsFunc: func(context.Context, *route53.ListResourceRecordSetsInput, ...func(*route53.Options)) (*route53.ListResourceRecordSetsOutput, error) {
				t.Fatal("ListResourceRecordSets must not be called when registrar hosted zone is absent")
				return nil, errUnexpectedMockRoute53Call
			},
			DeleteHostedZoneFunc: func(context.Context, *route53.DeleteHostedZoneInput, ...func(*route53.Options)) (*route53.DeleteHostedZoneOutput, error) {
				t.Fatal("DeleteHostedZone must not be called when registrar hosted zone is absent")
				return nil, errUnexpectedMockRoute53Call
			},
		},
	}

	schema := testDomainResourceSchema(t)
	req := resourceUpdateRequest(t, schema, plan, state)
	resp := &resource.UpdateResponse{State: tfsdk.State{Schema: schema}}

	domainResource.Update(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Update returned diagnostics: %v", resp.Diagnostics)
	}

	var got DomainRegistrationResourceModel
	resp.Diagnostics.Append(resp.State.Get(ctx, &got)...)
	if resp.Diagnostics.HasError() {
		t.Fatalf("reading response state returned diagnostics: %v", resp.Diagnostics)
	}
	if !got.DeleteHostedZone.ValueBool() {
		t.Fatal("delete_hosted_zone was not persisted")
	}
	if !got.HostedZoneID.IsNull() {
		t.Fatalf("hosted_zone_id = %q, want null", got.HostedZoneID.ValueString())
	}
}

func TestUpdateRetriesHostedZoneDeletionWhenStillTracked(t *testing.T) {
	ctx := context.Background()
	state := testDomainModel(t, "example.com")
	state.DeleteHostedZone = tftypes.BoolValue(true)
	state.HostedZoneID = tftypes.StringValue("ZREG")
	plan := state

	deleteCallCount := 0
	domainResource := &DomainRegistrationResource{
		client: &MockRoute53DomainsClient{
			GetDomainDetailFunc: func(_ context.Context, _ *route53domains.GetDomainDetailInput, _ ...func(*route53domains.Options)) (*route53domains.GetDomainDetailOutput, error) {
				return MockDomainDetailResponse("example.com"), nil
			},
		},
		route53Client: &MockRoute53Client{
			ListHostedZonesByNameFunc: func(context.Context, *route53.ListHostedZonesByNameInput, ...func(*route53.Options)) (*route53.ListHostedZonesByNameOutput, error) {
				return &route53.ListHostedZonesByNameOutput{
					HostedZones: []route53types.HostedZone{
						{
							Id:              aws.String("/hostedzone/ZREG"),
							Name:            aws.String("example.com."),
							CallerReference: aws.String("RISWorkflow-RD:test"),
							Config:          &route53types.HostedZoneConfig{Comment: aws.String(registrarHostedZoneComment)},
						},
					},
				}, nil
			},
			ListResourceRecordSetsFunc: func(context.Context, *route53.ListResourceRecordSetsInput, ...func(*route53.Options)) (*route53.ListResourceRecordSetsOutput, error) {
				return &route53.ListResourceRecordSetsOutput{
					ResourceRecordSets: []route53types.ResourceRecordSet{
						{Name: aws.String("example.com."), Type: route53types.RRTypeNs},
						{Name: aws.String("example.com."), Type: route53types.RRTypeSoa},
					},
				}, nil
			},
			DeleteHostedZoneFunc: func(context.Context, *route53.DeleteHostedZoneInput, ...func(*route53.Options)) (*route53.DeleteHostedZoneOutput, error) {
				deleteCallCount++
				return &route53.DeleteHostedZoneOutput{}, nil
			},
		},
	}

	schema := testDomainResourceSchema(t)
	req := resourceUpdateRequest(t, schema, plan, state)
	resp := &resource.UpdateResponse{State: tfsdk.State{Schema: schema}}

	domainResource.Update(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Update returned diagnostics: %v", resp.Diagnostics)
	}
	if deleteCallCount != 1 {
		t.Fatalf("DeleteHostedZone call count = %d, want %d", deleteCallCount, 1)
	}
}

func TestModifyPlanPlansHostedZoneCleanupRetry(t *testing.T) {
	ctx := context.Background()
	state := testDomainModel(t, "example.com")
	state.DeleteHostedZone = tftypes.BoolValue(true)
	state.HostedZoneID = tftypes.StringValue("ZREG")
	plan := state

	schema := testDomainResourceSchema(t)
	req := resource.ModifyPlanRequest{
		Plan:  tfsdk.Plan{Schema: schema},
		State: tfsdk.State{Schema: schema},
	}
	diags := req.Plan.Set(ctx, &plan)
	if diags.HasError() {
		t.Fatalf("setting modify plan returned diagnostics: %v", diags)
	}
	diags = req.State.Set(ctx, &state)
	if diags.HasError() {
		t.Fatalf("setting modify state returned diagnostics: %v", diags)
	}

	resp := &resource.ModifyPlanResponse{Plan: req.Plan}
	domainResource := &DomainRegistrationResource{}
	domainResource.ModifyPlan(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("ModifyPlan returned diagnostics: %v", resp.Diagnostics)
	}

	var got DomainRegistrationResourceModel
	resp.Diagnostics.Append(resp.Plan.Get(ctx, &got)...)
	if resp.Diagnostics.HasError() {
		t.Fatalf("reading modified plan returned diagnostics: %v", resp.Diagnostics)
	}
	if !got.HostedZoneID.IsNull() {
		t.Fatalf("planned hosted_zone_id = %q, want null", got.HostedZoneID.ValueString())
	}
}

func TestUpdateTagReconcilePreservesUnmanagedRemoteTags(t *testing.T) {
	ctx := context.Background()
	state := testDomainModel(t, "example.com")
	state.Tags = stringMapValue(t, map[string]string{
		"Environment": "prod",
		"OldManaged":  "remove",
	})
	state.TagsAll = stringMapValue(t, map[string]string{
		"Environment": "prod",
		"OldManaged":  "remove",
	})

	plan := state
	plan.Tags = stringMapValue(t, map[string]string{"Environment": "prod"})
	plan.TagsAll = stringMapValue(t, map[string]string{"Environment": "prod"})

	deletedTags := []string{}
	domainResource := &DomainRegistrationResource{
		client: &MockRoute53DomainsClient{
			ListTagsForDomainFunc: func(_ context.Context, _ *route53domains.ListTagsForDomainInput, _ ...func(*route53domains.Options)) (*route53domains.ListTagsForDomainOutput, error) {
				return &route53domains.ListTagsForDomainOutput{
					TagList: []types.Tag{
						{Key: aws.String("Environment"), Value: aws.String("prod")},
						{Key: aws.String("OldManaged"), Value: aws.String("remove")},
						{Key: aws.String("External"), Value: aws.String("console")},
					},
				}, nil
			},
			DeleteTagsForDomainFunc: func(_ context.Context, params *route53domains.DeleteTagsForDomainInput, _ ...func(*route53domains.Options)) (*route53domains.DeleteTagsForDomainOutput, error) {
				deletedTags = append(deletedTags, params.TagsToDelete...)
				return &route53domains.DeleteTagsForDomainOutput{}, nil
			},
			UpdateTagsForDomainFunc: func(context.Context, *route53domains.UpdateTagsForDomainInput, ...func(*route53domains.Options)) (*route53domains.UpdateTagsForDomainOutput, error) {
				t.Fatal("UpdateTagsForDomain must not be called when desired managed tags already match")
				return nil, errUnexpectedMockRoute53DomainsCall
			},
			GetDomainDetailFunc: func(_ context.Context, _ *route53domains.GetDomainDetailInput, _ ...func(*route53domains.Options)) (*route53domains.GetDomainDetailOutput, error) {
				return MockDomainDetailResponse("example.com"), nil
			},
		},
		route53Client: &MockRoute53Client{
			ListHostedZonesByNameFunc: func(context.Context, *route53.ListHostedZonesByNameInput, ...func(*route53.Options)) (*route53.ListHostedZonesByNameOutput, error) {
				return &route53.ListHostedZonesByNameOutput{}, nil
			},
		},
	}

	schema := testDomainResourceSchema(t)
	req := resourceUpdateRequest(t, schema, plan, state)
	resp := &resource.UpdateResponse{State: tfsdk.State{Schema: schema}}

	domainResource.Update(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Update returned diagnostics: %v", resp.Diagnostics)
	}
	if !reflect.DeepEqual(deletedTags, []string{"OldManaged"}) {
		t.Fatalf("deleted tags = %#v, want %#v", deletedTags, []string{"OldManaged"})
	}
}

func TestDeleteWaitsForDomainDeletionOperation(t *testing.T) {
	ctx := context.Background()
	state := testDomainModel(t, "example.com")
	state.AllowDelete = tftypes.BoolValue(true)

	waitedForOperation := false
	domainResource := &DomainRegistrationResource{
		client: &MockRoute53DomainsClient{
			DeleteDomainFunc: func(_ context.Context, _ *route53domains.DeleteDomainInput, _ ...func(*route53domains.Options)) (*route53domains.DeleteDomainOutput, error) {
				return &route53domains.DeleteDomainOutput{OperationId: aws.String("op-delete")}, nil
			},
			GetOperationDetailFunc: func(_ context.Context, params *route53domains.GetOperationDetailInput, _ ...func(*route53domains.Options)) (*route53domains.GetOperationDetailOutput, error) {
				if got := aws.ToString(params.OperationId); got != "op-delete" {
					t.Fatalf("operation ID = %q, want %q", got, "op-delete")
				}
				waitedForOperation = true
				return &route53domains.GetOperationDetailOutput{Status: types.OperationStatusSuccessful}, nil
			},
		},
		route53Client: &MockRoute53Client{
			ListHostedZonesByNameFunc: func(context.Context, *route53.ListHostedZonesByNameInput, ...func(*route53.Options)) (*route53.ListHostedZonesByNameOutput, error) {
				return &route53.ListHostedZonesByNameOutput{}, nil
			},
		},
	}

	schema := testDomainResourceSchema(t)
	req := resourceDeleteRequest(t, schema, state)
	resp := &resource.DeleteResponse{}

	domainResource.Delete(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Delete returned diagnostics: %v", resp.Diagnostics)
	}
	if !waitedForOperation {
		t.Fatal("GetOperationDetail was not called for domain deletion")
	}
}

func TestDeleteReturnsErrorWhenDomainDeletionOperationFails(t *testing.T) {
	ctx := context.Background()
	state := testDomainModel(t, "example.com")
	state.AllowDelete = tftypes.BoolValue(true)

	domainResource := &DomainRegistrationResource{
		client: &MockRoute53DomainsClient{
			DeleteDomainFunc: func(_ context.Context, _ *route53domains.DeleteDomainInput, _ ...func(*route53domains.Options)) (*route53domains.DeleteDomainOutput, error) {
				return &route53domains.DeleteDomainOutput{OperationId: aws.String("op-delete")}, nil
			},
			GetOperationDetailFunc: func(_ context.Context, _ *route53domains.GetOperationDetailInput, _ ...func(*route53domains.Options)) (*route53domains.GetOperationDetailOutput, error) {
				return &route53domains.GetOperationDetailOutput{
					Status:  types.OperationStatusFailed,
					Message: aws.String("registry refused deletion"),
				}, nil
			},
		},
	}

	schema := testDomainResourceSchema(t)
	req := resourceDeleteRequest(t, schema, state)
	resp := &resource.DeleteResponse{}

	domainResource.Delete(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected domain deletion failure diagnostic")
	}
}

func stringListValue(t *testing.T, values ...string) tftypes.List {
	t.Helper()

	list, diags := tftypes.ListValueFrom(context.Background(), tftypes.StringType, values)
	if diags.HasError() {
		t.Fatalf("creating string list returned diagnostics: %v", diags)
	}
	return list
}

func testContactModel(email string) *ContactModel {
	return &ContactModel{
		FirstName:    stringValue("John"),
		LastName:     stringValue("Doe"),
		Email:        stringValue(email),
		PhoneNumber:  stringValue("+1.5551234567"),
		AddressLine1: stringValue("123 Main St"),
		AddressLine2: tftypes.StringNull(),
		City:         stringValue("Seattle"),
		State:        stringValue("WA"),
		ZipCode:      stringValue("98101"),
		CountryCode:  stringValue("US"),
		ContactType:  stringValue("PERSON"),
	}
}

func stringMapValue(t *testing.T, values map[string]string) tftypes.Map {
	t.Helper()

	result, diags := stringMapToFrameworkMap(values)
	if diags.HasError() {
		t.Fatalf("creating string map returned diagnostics: %v", diags)
	}
	return result
}

func testDomainResourceSchema(t *testing.T) resourceschema.Schema {
	t.Helper()

	ctx := context.Background()
	r := NewDomainRegistrationResource()
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema returned diagnostics: %v", resp.Diagnostics)
	}
	return resp.Schema
}

func testDomainModel(t *testing.T, domainName string) DomainRegistrationResourceModel {
	t.Helper()

	return DomainRegistrationResourceModel{
		ID:                      tftypes.StringValue(domainName),
		DomainName:              tftypes.StringValue(domainName),
		DurationYears:           tftypes.Int64Value(1),
		AutoRenew:               tftypes.BoolValue(false),
		AdminContact:            testContactModel("admin@example.com"),
		RegistrantContact:       testContactModel("registrant@example.com"),
		TechContact:             testContactModel("tech@example.com"),
		AdminPrivacy:            tftypes.BoolValue(true),
		RegistrantPrivacy:       tftypes.BoolValue(true),
		TechPrivacy:             tftypes.BoolValue(true),
		Nameservers:             stringListValue(t, "ns1.example.com", "ns2.example.com"),
		Tags:                    emptyFrameworkStringMap(),
		TagsAll:                 emptyFrameworkStringMap(),
		AllowDelete:             tftypes.BoolValue(false),
		DeleteHostedZone:        tftypes.BoolValue(false),
		Status:                  tftypes.StringValue("ok"),
		ExpirationDate:          tftypes.StringValue(time.Now().AddDate(1, 0, 0).Format(time.RFC3339)),
		CreationDate:            tftypes.StringValue(time.Now().Format(time.RFC3339)),
		RegistrationTimeout:     tftypes.Int64Value(900),
		RegistrationOperationID: tftypes.StringNull(),
		HostedZoneID:            tftypes.StringNull(),
	}
}

func resourceCreateRequest(t *testing.T, schema resourceschema.Schema, plan DomainRegistrationResourceModel) resource.CreateRequest {
	t.Helper()

	req := resource.CreateRequest{Plan: tfsdk.Plan{Schema: schema}}
	diags := req.Plan.Set(context.Background(), &plan)
	if diags.HasError() {
		t.Fatalf("setting create plan returned diagnostics: %v", diags)
	}
	return req
}

func resourceReadRequest(t *testing.T, schema resourceschema.Schema, state DomainRegistrationResourceModel) resource.ReadRequest {
	t.Helper()

	req := resource.ReadRequest{State: tfsdk.State{Schema: schema}}
	diags := req.State.Set(context.Background(), &state)
	if diags.HasError() {
		t.Fatalf("setting read state returned diagnostics: %v", diags)
	}
	return req
}

func resourceUpdateRequest(t *testing.T, schema resourceschema.Schema, plan, state DomainRegistrationResourceModel) resource.UpdateRequest {
	t.Helper()

	req := resource.UpdateRequest{
		Plan:  tfsdk.Plan{Schema: schema},
		State: tfsdk.State{Schema: schema},
	}
	diags := req.Plan.Set(context.Background(), &plan)
	if diags.HasError() {
		t.Fatalf("setting update plan returned diagnostics: %v", diags)
	}
	diags = req.State.Set(context.Background(), &state)
	if diags.HasError() {
		t.Fatalf("setting update state returned diagnostics: %v", diags)
	}
	return req
}

func resourceDeleteRequest(t *testing.T, schema resourceschema.Schema, state DomainRegistrationResourceModel) resource.DeleteRequest {
	t.Helper()

	req := resource.DeleteRequest{State: tfsdk.State{Schema: schema}}
	diags := req.State.Set(context.Background(), &state)
	if diags.HasError() {
		t.Fatalf("setting delete state returned diagnostics: %v", diags)
	}
	return req
}

// Helper to create terraform string values for testing
func stringValue(s string) tftypes.String {
	return tftypes.StringValue(s)
}

func TestSelectRegistrarHostedZone(t *testing.T) {
	tests := []struct {
		name        string
		domainName  string
		zones       []route53types.HostedZone
		wantID      string
		expectError bool
	}{
		{
			name:       "selects registrar zone among exact matches",
			domainName: "kafanpadaxri.com",
			zones: []route53types.HostedZone{
				{
					Id:              aws.String("/hostedzone/ZTF"),
					Name:            aws.String("kafanpadaxri.com."),
					CallerReference: aws.String("terraform-20260423033925532200000001"),
					Config:          &route53types.HostedZoneConfig{Comment: aws.String("Managed by Terraform")},
				},
				{
					Id:              aws.String("/hostedzone/ZREG"),
					Name:            aws.String("kafanpadaxri.com."),
					CallerReference: aws.String("RISWorkflow-RD:ebaf2de3-0d22-4e7d-80d8-e242984722ac"),
					Config:          &route53types.HostedZoneConfig{Comment: aws.String(registrarHostedZoneComment)},
				},
			},
			wantID: "/hostedzone/ZREG",
		},
		{
			name:       "ignores non-exact and private zones",
			domainName: "kafanpadaxri.com",
			zones: []route53types.HostedZone{
				{
					Id:              aws.String("/hostedzone/ZPRIVATE"),
					Name:            aws.String("kafanpadaxri.com."),
					CallerReference: aws.String("RISWorkflow-RD:private"),
					Config:          &route53types.HostedZoneConfig{PrivateZone: true},
				},
				{
					Id:              aws.String("/hostedzone/ZOTHER"),
					Name:            aws.String("sub.kafanpadaxri.com."),
					CallerReference: aws.String("RISWorkflow-RD:other"),
					Config:          &route53types.HostedZoneConfig{Comment: aws.String(registrarHostedZoneComment)},
				},
				{
					Id:              aws.String("/hostedzone/ZREG"),
					Name:            aws.String("kafanpadaxri.com."),
					CallerReference: aws.String("RISWorkflow-RD:real"),
					Config:          &route53types.HostedZoneConfig{Comment: aws.String(registrarHostedZoneComment)},
				},
			},
			wantID: "/hostedzone/ZREG",
		},
		{
			name:       "errors when registrar zone missing",
			domainName: "kafanpadaxri.com",
			zones: []route53types.HostedZone{
				{
					Id:              aws.String("/hostedzone/ZTF"),
					Name:            aws.String("kafanpadaxri.com."),
					CallerReference: aws.String("terraform-20260423033925532200000001"),
					Config:          &route53types.HostedZoneConfig{Comment: aws.String("Managed by Terraform")},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zone, err := selectRegistrarHostedZone(tt.domainName, tt.zones)
			if tt.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := aws.ToString(zone.Id); got != tt.wantID {
				t.Fatalf("selected zone ID %q, want %q", got, tt.wantID)
			}
		})
	}
}

func TestListExactHostedZonesAcrossPages(t *testing.T) {
	resource := &DomainRegistrationResource{}
	callCount := 0
	resource.route53Client = &MockRoute53Client{
		ListHostedZonesByNameFunc: func(_ context.Context, params *route53.ListHostedZonesByNameInput, _ ...func(*route53.Options)) (*route53.ListHostedZonesByNameOutput, error) {
			callCount++

			switch callCount {
			case 1:
				if got := aws.ToString(params.DNSName); got != "example.com" {
					t.Fatalf("first call DNSName = %q, want %q", got, "example.com")
				}
				if params.HostedZoneId != nil {
					t.Fatalf("first call HostedZoneId = %q, want nil", aws.ToString(params.HostedZoneId))
				}

				return &route53.ListHostedZonesByNameOutput{
					HostedZones: []route53types.HostedZone{
						{Id: aws.String("/hostedzone/Z1"), Name: aws.String("example.com.")},
					},
					IsTruncated:      true,
					NextDNSName:      aws.String("example.com"),
					NextHostedZoneId: aws.String("Z2"),
				}, nil
			case 2:
				if got := aws.ToString(params.DNSName); got != "example.com" {
					t.Fatalf("second call DNSName = %q, want %q", got, "example.com")
				}
				if got := aws.ToString(params.HostedZoneId); got != "Z2" {
					t.Fatalf("second call HostedZoneId = %q, want %q", got, "Z2")
				}

				return &route53.ListHostedZonesByNameOutput{
					HostedZones: []route53types.HostedZone{
						{Id: aws.String("/hostedzone/Z2"), Name: aws.String("example.com.")},
						{Id: aws.String("/hostedzone/Z3"), Name: aws.String("example.net.")},
					},
					IsTruncated: false,
				}, nil
			default:
				t.Fatalf("unexpected extra ListHostedZonesByName call %d", callCount)
				return nil, errUnexpectedMockRoute53Call
			}
		},
	}

	zones, err := resource.listExactHostedZones(context.Background(), "example.com")
	if err != nil {
		t.Fatalf("listExactHostedZones returned error: %v", err)
	}
	if callCount != 2 {
		t.Fatalf("ListHostedZonesByName call count = %d, want %d", callCount, 2)
	}
	if len(zones) != 2 {
		t.Fatalf("zone count = %d, want %d", len(zones), 2)
	}
}

func TestListExactHostedZonesStopsAfterFirstNonMatch(t *testing.T) {
	resource := &DomainRegistrationResource{}
	callCount := 0
	resource.route53Client = &MockRoute53Client{
		ListHostedZonesByNameFunc: func(_ context.Context, _ *route53.ListHostedZonesByNameInput, _ ...func(*route53.Options)) (*route53.ListHostedZonesByNameOutput, error) {
			callCount++
			return &route53.ListHostedZonesByNameOutput{
				HostedZones: []route53types.HostedZone{
					{Id: aws.String("/hostedzone/Z9"), Name: aws.String("example.net.")},
				},
				IsTruncated:      true,
				NextDNSName:      aws.String("example.org"),
				NextHostedZoneId: aws.String("Z10"),
			}, nil
		},
	}

	zones, err := resource.listExactHostedZones(context.Background(), "example.com")
	if err != nil {
		t.Fatalf("listExactHostedZones returned error: %v", err)
	}
	if callCount != 1 {
		t.Fatalf("ListHostedZonesByName call count = %d, want %d", callCount, 1)
	}
	if len(zones) != 0 {
		t.Fatalf("zone count = %d, want %d", len(zones), 0)
	}
}

func TestDeleteRegistrarHostedZoneRejectsCustomRecordOnLaterPage(t *testing.T) {
	resource := &DomainRegistrationResource{}
	recordSetCallCount := 0
	deleteCallCount := 0
	resource.route53Client = &MockRoute53Client{
		ListHostedZonesByNameFunc: func(_ context.Context, _ *route53.ListHostedZonesByNameInput, _ ...func(*route53.Options)) (*route53.ListHostedZonesByNameOutput, error) {
			return &route53.ListHostedZonesByNameOutput{
				HostedZones: []route53types.HostedZone{
					{
						Id:              aws.String("/hostedzone/ZREG"),
						Name:            aws.String("example.com."),
						CallerReference: aws.String("RISWorkflow-RD:test"),
						Config:          &route53types.HostedZoneConfig{Comment: aws.String(registrarHostedZoneComment)},
					},
				},
			}, nil
		},
		ListResourceRecordSetsFunc: func(_ context.Context, params *route53.ListResourceRecordSetsInput, _ ...func(*route53.Options)) (*route53.ListResourceRecordSetsOutput, error) {
			recordSetCallCount++

			switch recordSetCallCount {
			case 1:
				if got := aws.ToString(params.HostedZoneId); got != "/hostedzone/ZREG" {
					t.Fatalf("first call HostedZoneId = %q, want %q", got, "/hostedzone/ZREG")
				}
				if params.StartRecordName != nil {
					t.Fatalf("first call StartRecordName = %q, want nil", aws.ToString(params.StartRecordName))
				}
				return &route53.ListResourceRecordSetsOutput{
					ResourceRecordSets: []route53types.ResourceRecordSet{
						{Name: aws.String("example.com."), Type: route53types.RRTypeNs},
						{Name: aws.String("example.com."), Type: route53types.RRTypeSoa},
					},
					IsTruncated:          true,
					NextRecordName:       aws.String("www.example.com."),
					NextRecordType:       route53types.RRTypeA,
					NextRecordIdentifier: aws.String("weighted"),
				}, nil
			case 2:
				if got := aws.ToString(params.StartRecordName); got != "www.example.com." {
					t.Fatalf("second call StartRecordName = %q, want %q", got, "www.example.com.")
				}
				if params.StartRecordType != route53types.RRTypeA {
					t.Fatalf("second call StartRecordType = %q, want %q", params.StartRecordType, route53types.RRTypeA)
				}
				if got := aws.ToString(params.StartRecordIdentifier); got != "weighted" {
					t.Fatalf("second call StartRecordIdentifier = %q, want %q", got, "weighted")
				}
				return &route53.ListResourceRecordSetsOutput{
					ResourceRecordSets: []route53types.ResourceRecordSet{
						{Name: aws.String("www.example.com."), Type: route53types.RRTypeA},
					},
					IsTruncated: false,
				}, nil
			default:
				t.Fatalf("unexpected extra ListResourceRecordSets call %d", recordSetCallCount)
				return nil, errUnexpectedMockRoute53Call
			}
		},
		DeleteHostedZoneFunc: func(_ context.Context, _ *route53.DeleteHostedZoneInput, _ ...func(*route53.Options)) (*route53.DeleteHostedZoneOutput, error) {
			deleteCallCount++
			return &route53.DeleteHostedZoneOutput{}, nil
		},
	}

	err := resource.deleteRegistrarHostedZone(context.Background(), "example.com")
	if err == nil {
		t.Fatal("expected deleteRegistrarHostedZone to return an error")
	}
	if !strings.Contains(err.Error(), "custom record") {
		t.Fatalf("error = %q, want custom record failure", err)
	}
	if recordSetCallCount != 2 {
		t.Fatalf("ListResourceRecordSets call count = %d, want %d", recordSetCallCount, 2)
	}
	if deleteCallCount != 0 {
		t.Fatalf("DeleteHostedZone call count = %d, want %d", deleteCallCount, 0)
	}
}

// MockDomainDetailResponse creates a mock GetDomainDetailOutput
func MockDomainDetailResponse(domainName string) *route53domains.GetDomainDetailOutput {
	now := time.Now()
	expiry := now.AddDate(1, 0, 0)
	return &route53domains.GetDomainDetailOutput{
		DomainName:     aws.String(domainName),
		AutoRenew:      aws.Bool(false),
		CreationDate:   aws.Time(now),
		ExpirationDate: aws.Time(expiry),
		StatusList:     []string{"ok"},
		Nameservers: []types.Nameserver{
			{Name: aws.String("ns1.example.com")},
			{Name: aws.String("ns2.example.com")},
		},
		AdminContact: &types.ContactDetail{
			FirstName:   aws.String("John"),
			LastName:    aws.String("Doe"),
			Email:       aws.String("admin@example.com"),
			PhoneNumber: aws.String("+1.5551234567"),
		},
		RegistrantContact: &types.ContactDetail{
			FirstName:   aws.String("John"),
			LastName:    aws.String("Doe"),
			Email:       aws.String("registrant@example.com"),
			PhoneNumber: aws.String("+1.5551234567"),
		},
		TechContact: &types.ContactDetail{
			FirstName:   aws.String("John"),
			LastName:    aws.String("Doe"),
			Email:       aws.String("tech@example.com"),
			PhoneNumber: aws.String("+1.5551234567"),
		},
		AdminPrivacy:      aws.Bool(true),
		RegistrantPrivacy: aws.Bool(true),
		TechPrivacy:       aws.Bool(true),
	}
}
