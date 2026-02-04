package provider

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
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

// Helper to create terraform string values for testing
func stringValue(s string) tftypes.String {
	return tftypes.StringValue(s)
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
