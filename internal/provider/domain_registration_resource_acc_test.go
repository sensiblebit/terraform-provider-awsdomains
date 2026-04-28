package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDomainRegistration_planOnlyOmittedNameservers(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             testAccDomainRegistrationConfig("example-plan-only.invalid", "", map[string]string{"Environment": "test"}),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDomainRegistration_planOnlyCustomNameserversAndTags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainRegistrationConfig(
					"example-plan-only.invalid",
					`
  nameservers = [
    "ns1.example.com",
    "ns2.example.com",
  ]
`,
					map[string]string{"Environment": "test", "Name": "example-plan-only.invalid"},
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccDomainRegistrationConfig(domainName, nameserversConfig string, tags map[string]string) string {
	return fmt.Sprintf(`
provider "awsdomains" {
  region = "us-east-1"

  default_tags {
    tags = {
      ManagedBy = "terraform"
    }
  }
}

resource "awsdomains_domain" "test" {
  domain_name = %[1]q

  admin_contact = {
    first_name     = "Test"
    last_name      = "User"
    email          = "noreply@example.com"
    phone_number   = "+1.5551234567"
    address_line_1 = "123 Test St"
    city           = "Seattle"
    state          = "WA"
    zip_code       = "98101"
    country_code   = "US"
  }

  registrant_contact = {
    first_name     = "Test"
    last_name      = "User"
    email          = "noreply@example.com"
    phone_number   = "+1.5551234567"
    address_line_1 = "123 Test St"
    city           = "Seattle"
    state          = "WA"
    zip_code       = "98101"
    country_code   = "US"
  }

  tech_contact = {
    first_name     = "Test"
    last_name      = "User"
    email          = "noreply@example.com"
    phone_number   = "+1.5551234567"
    address_line_1 = "123 Test St"
    city           = "Seattle"
    state          = "WA"
    zip_code       = "98101"
    country_code   = "US"
  }
%[2]s
  tags = %[3]s
}
`, domainName, nameserversConfig, testAccTagsConfig(tags))
}

func testAccTagsConfig(tags map[string]string) string {
	if len(tags) == 0 {
		return "{}"
	}

	var result strings.Builder
	result.WriteString("{\n")
	for key, value := range tags {
		fmt.Fprintf(&result, "    %s = %q\n", key, value)
	}
	result.WriteString("  }")
	return result.String()
}
