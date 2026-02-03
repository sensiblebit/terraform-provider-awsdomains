---
page_title: "awsdomains_domain_availability Data Source - terraform-provider-awsdomains"
subcategory: ""
description: |-
  Check if a domain name is available for registration.
---

# awsdomains_domain_availability (Data Source)

Check if a domain name is available for registration. This is a free API call with no cost.

## Example Usage

```terraform
data "awsdomains_domain_availability" "check" {
  domain_name = "example.com"
}

output "is_available" {
  value = data.awsdomains_domain_availability.check.available
}

output "availability_status" {
  value = data.awsdomains_domain_availability.check.availability
}
```

## Schema

### Required

- `domain_name` (String) The domain name to check availability for.

### Read-Only

- `id` (String) The domain name.
- `availability` (String) Availability status. One of: `AVAILABLE`, `AVAILABLE_RESERVED`, `AVAILABLE_PREORDER`, `UNAVAILABLE`, `UNAVAILABLE_PREMIUM`, `UNAVAILABLE_RESTRICTED`, `RESERVED`, `DONT_KNOW`.
- `available` (Boolean) `true` if the domain can be registered, `false` otherwise.
