package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDomainPriceDataSource_com(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainPriceDataSourceConfig("com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.awsdomains_domain_price.test", "tld", "com"),
					resource.TestCheckResourceAttrSet("data.awsdomains_domain_price.test", "registration_price"),
					resource.TestCheckResourceAttrSet("data.awsdomains_domain_price.test", "renewal_price"),
					resource.TestCheckResourceAttr("data.awsdomains_domain_price.test", "currency", "USD"),
				),
			},
		},
	})
}

func TestAccDomainPriceDataSource_net(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainPriceDataSourceConfig("net"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.awsdomains_domain_price.test", "tld", "net"),
					resource.TestCheckResourceAttrSet("data.awsdomains_domain_price.test", "registration_price"),
				),
			},
		},
	})
}

func testAccDomainPriceDataSourceConfig(tld string) string {
	return `
provider "awsdomains" {
  region = "us-east-1"
}

data "awsdomains_domain_price" "test" {
  tld = "` + tld + `"
}
`
}
