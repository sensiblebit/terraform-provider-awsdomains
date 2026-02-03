package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDomainAvailabilityDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainAvailabilityDataSourceConfig("google.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.awsdomains_domain_availability.test", "domain_name", "google.com"),
					resource.TestCheckResourceAttr("data.awsdomains_domain_availability.test", "availability", "UNAVAILABLE"),
					resource.TestCheckResourceAttr("data.awsdomains_domain_availability.test", "available", "false"),
				),
			},
		},
	})
}

func TestAccDomainAvailabilityDataSource_availableDomain(t *testing.T) {
	// Test with a domain that's likely to be available
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainAvailabilityDataSourceConfig("xyzzy-test-domain-12345678.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.awsdomains_domain_availability.test", "domain_name", "xyzzy-test-domain-12345678.com"),
					// Just check that it returns a valid availability status
					resource.TestCheckResourceAttrSet("data.awsdomains_domain_availability.test", "availability"),
					resource.TestCheckResourceAttrSet("data.awsdomains_domain_availability.test", "available"),
				),
			},
		},
	})
}

func testAccDomainAvailabilityDataSourceConfig(domain string) string {
	return `
provider "awsdomains" {
  region = "us-east-1"
}

data "awsdomains_domain_availability" "test" {
  domain_name = "` + domain + `"
}
`
}
