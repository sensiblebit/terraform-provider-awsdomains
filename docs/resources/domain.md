---
page_title: "awsdomains_domain Resource - terraform-provider-awsdomains"
subcategory: ""
description: |-
  Registers and manages a domain name through AWS Route53 Domains.
---

# awsdomains_domain (Resource)

Registers and manages a domain name through AWS Route53 Domains.

~> **Important:** Domain registration incurs costs ($12-35+ depending on TLD). By default, domains are NOT deleted on `terraform destroy` to prevent accidental loss. Set `allow_delete = true` to enable actual deletion (also deletes the hosted zone if empty).

## Example Usage

### Basic Registration

```terraform
resource "awsdomains_domain" "example" {
  domain_name    = "example.com"
  duration_years = 1
  auto_renew     = false

  admin_contact = {
    first_name     = "John"
    last_name      = "Doe"
    email          = "admin@example.com"
    phone_number   = "+1.5551234567"
    address_line_1 = "123 Main St"
    city           = "Seattle"
    state          = "WA"
    zip_code       = "98101"
    country_code   = "US"
    contact_type   = "PERSON"
  }

  registrant_contact = {
    first_name     = "John"
    last_name      = "Doe"
    email          = "registrant@example.com"
    phone_number   = "+1.5551234567"
    address_line_1 = "123 Main St"
    city           = "Seattle"
    state          = "WA"
    zip_code       = "98101"
    country_code   = "US"
    contact_type   = "PERSON"
  }

  tech_contact = {
    first_name     = "John"
    last_name      = "Doe"
    email          = "tech@example.com"
    phone_number   = "+1.5551234567"
    address_line_1 = "123 Main St"
    city           = "Seattle"
    state          = "WA"
    zip_code       = "98101"
    country_code   = "US"
    contact_type   = "PERSON"
  }

  admin_privacy      = true
  registrant_privacy = true
  tech_privacy       = true
}
```

### Using the Hosted Zone

AWS automatically creates a Route53 hosted zone when registering a domain. The `hosted_zone_id` attribute provides direct access:

```terraform
resource "awsdomains_domain" "example" {
  domain_name = "example.com"
  # ... contacts ...
}

# Create DNS records without needing a data source lookup
resource "aws_route53_record" "www" {
  zone_id = awsdomains_domain.example.hosted_zone_id
  name    = "www.example.com"
  type    = "A"
  ttl     = 300
  records = ["192.0.2.1"]
}

# Use with ACME/Let's Encrypt
resource "acme_certificate" "cert" {
  # ...
  dns_challenge {
    provider = "route53"
    config = {
      AWS_HOSTED_ZONE_ID = awsdomains_domain.example.hosted_zone_id
    }
  }
}
```

### With Custom Nameservers

```terraform
resource "awsdomains_domain" "example" {
  domain_name = "example.com"
  # ... contacts ...

  nameservers = [
    "ns1.example.com",
    "ns2.example.com",
  ]
}
```

## Schema

### Required

- `domain_name` (String) The domain name to register. Cannot be changed after creation.
- `admin_contact` (Attributes) Administrative contact details. See [Contact](#nestedatt--contact) below.
- `registrant_contact` (Attributes) Registrant contact details. See [Contact](#nestedatt--contact) below.
- `tech_contact` (Attributes) Technical contact details. See [Contact](#nestedatt--contact) below.

### Optional

- `duration_years` (Number) Number of years to register the domain (1-10). Defaults to `1`.
- `auto_renew` (Boolean) Whether to enable automatic renewal. Defaults to `false`.
- `admin_privacy` (Boolean) Enable WHOIS privacy for admin contact. Defaults to `true`.
- `registrant_privacy` (Boolean) Enable WHOIS privacy for registrant contact. Defaults to `true`.
- `tech_privacy` (Boolean) Enable WHOIS privacy for tech contact. Defaults to `true`.
- `nameservers` (List of String) Custom nameservers for the domain.
- `allow_delete` (Boolean) Allow actual domain deletion on `terraform destroy`. Defaults to `false`.
- `registration_timeout` (Number) Timeout in seconds for domain registration. Defaults to `900`.

### Read-Only

- `id` (String) The domain name.
- `status` (String) Current status of the domain.
- `creation_date` (String) Domain creation date in RFC3339 format.
- `expiration_date` (String) Domain expiration date in RFC3339 format.
- `hosted_zone_id` (String) The Route53 hosted zone ID automatically created for this domain.

<a id="nestedatt--contact"></a>
### Contact

Required:

- `first_name` (String) First name.
- `last_name` (String) Last name.
- `email` (String) Email address.
- `phone_number` (String) Phone number in E.164 format (e.g., `+1.5551234567`).
- `address_line_1` (String) Street address line 1.
- `city` (String) City.
- `state` (String) State or province.
- `zip_code` (String) Postal code.
- `country_code` (String) Two-letter country code (e.g., `US`, `UK`).

Optional:

- `address_line_2` (String) Street address line 2.
- `contact_type` (String) Contact type: `PERSON`, `COMPANY`, `ASSOCIATION`, `PUBLIC_BODY`, or `RESELLER`. Defaults to `PERSON`.

## Import

Domains can be imported using the domain name:

```shell
terraform import awsdomains_domain.example example.com
```

~> **Note:** Contact information is not populated during import. After importing, run `terraform apply` to set contact details from your configuration.
