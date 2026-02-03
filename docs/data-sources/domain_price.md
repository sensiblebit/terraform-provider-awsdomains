---
page_title: "awsdomains_domain_price Data Source - terraform-provider-awsdomains"
subcategory: ""
description: |-
  Get pricing information for a top-level domain (TLD).
---

# awsdomains_domain_price (Data Source)

Get pricing information for a top-level domain (TLD). This is a free API call with no cost.

## Example Usage

```terraform
data "awsdomains_domain_price" "com" {
  tld = "com"
}

output "registration_cost" {
  value = "${data.awsdomains_domain_price.com.registration_price} ${data.awsdomains_domain_price.com.currency}"
}

output "renewal_cost" {
  value = "${data.awsdomains_domain_price.com.renewal_price} ${data.awsdomains_domain_price.com.currency}"
}
```

## Schema

### Required

- `tld` (String) The top-level domain to get pricing for (e.g., `com`, `net`, `org`).

### Read-Only

- `id` (String) The TLD.
- `registration_price` (Number) Cost to register a domain with this TLD.
- `renewal_price` (Number) Cost to renew a domain with this TLD.
- `transfer_price` (Number) Cost to transfer a domain with this TLD.
- `restoration_price` (Number) Cost to restore an expired domain with this TLD.
- `currency` (String) Currency code (typically `USD`).
