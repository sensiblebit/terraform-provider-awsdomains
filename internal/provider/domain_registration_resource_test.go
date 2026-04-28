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
	tftypes "github.com/hashicorp/terraform-plugin-framework/types"
)

// MockRoute53DomainsClient is a mock implementation for testing
type MockRoute53DomainsClient struct {
	GetDomainDetailFunc         func(ctx context.Context, params *route53domains.GetDomainDetailInput, optFns ...func(*route53domains.Options)) (*route53domains.GetDomainDetailOutput, error)
	RegisterDomainFunc          func(ctx context.Context, params *route53domains.RegisterDomainInput, optFns ...func(*route53domains.Options)) (*route53domains.RegisterDomainOutput, error)
	GetOperationDetailFunc      func(ctx context.Context, params *route53domains.GetOperationDetailInput, optFns ...func(*route53domains.Options)) (*route53domains.GetOperationDetailOutput, error)
	UpdateDomainNameserversFunc func(ctx context.Context, params *route53domains.UpdateDomainNameserversInput, optFns ...func(*route53domains.Options)) (*route53domains.UpdateDomainNameserversOutput, error)
	CheckDomainAvailabilityFunc func(ctx context.Context, params *route53domains.CheckDomainAvailabilityInput, optFns ...func(*route53domains.Options)) (*route53domains.CheckDomainAvailabilityOutput, error)
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
)

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
		"hosted_zone_id",
	}

	for _, attr := range requiredAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema missing '%s' attribute", attr)
		}
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
