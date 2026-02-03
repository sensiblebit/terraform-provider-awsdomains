# Terraform Provider for AWS Route53 Domain Registration

A Terraform provider for registering and managing domain names through AWS Route53 Domains.

## Features

- Register new domains with AWS Route53 Domains
- Manage domain contacts (admin, registrant, tech)
- Configure WHOIS privacy protection
- Update nameservers
- Manage auto-renewal settings
- Import existing domains into Terraform state
- Safe defaults: domains are NOT deleted on `terraform destroy` unless explicitly enabled

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (for building from source)
- AWS credentials with Route53 Domains permissions

## Installation

### From Terraform Registry

```hcl
terraform {
  required_providers {
    awsdomains = {
      source  = "sensiblebit/awsdomains"
      version = "~> 1.0"
    }
  }
}
```

### Building from Source

```bash
git clone https://github.com/sensiblebit/terraform-provider-awsdomains.git
cd terraform-provider-awsdomains
go build -o terraform-provider-awsdomains
```

## Usage

### Provider Configuration

```hcl
provider "awsdomains" {
  region  = "us-east-1"  # Required: Route53 Domains only works in us-east-1
  profile = "default"    # Optional: AWS profile name
}
```

### Register a Domain

```hcl
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

  # Optional: Set custom nameservers
  nameservers = [
    "ns1.example.com",
    "ns2.example.com",
  ]

  # Safety: Set to true to allow actual domain deletion
  allow_delete = false
}
```

### Import Existing Domain

```bash
terraform import 'awsdomains_domain.example' example.com
```

## Data Sources

### awsdomains_domain_availability

Check if a domain is available for registration (free API - no cost).

```hcl
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

**Attributes:**
| Name | Type | Description |
|------|------|-------------|
| `domain_name` | string | The domain to check |
| `availability` | string | Status: AVAILABLE, UNAVAILABLE, etc. |
| `available` | bool | True if domain can be registered |

### awsdomains_domain_price

Get pricing for a TLD (free API - no cost).

```hcl
data "awsdomains_domain_price" "com" {
  tld = "com"
}

output "registration_cost" {
  value = "${data.awsdomains_domain_price.com.registration_price} ${data.awsdomains_domain_price.com.currency}"
}
```

**Attributes:**
| Name | Type | Description |
|------|------|-------------|
| `tld` | string | Top-level domain (com, net, org, etc.) |
| `registration_price` | number | Cost to register |
| `renewal_price` | number | Cost to renew |
| `transfer_price` | number | Cost to transfer |
| `currency` | string | Currency code (USD) |

## Resource: awsdomains_domain

### Arguments

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `domain_name` | string | Yes | The domain name to register |
| `duration_years` | number | No | Years to register (1-10, default: 1) |
| `auto_renew` | bool | No | Enable auto-renewal (default: false) |
| `admin_contact` | object | Yes | Administrative contact details |
| `registrant_contact` | object | Yes | Registrant contact details |
| `tech_contact` | object | Yes | Technical contact details |
| `admin_privacy` | bool | No | WHOIS privacy for admin (default: true) |
| `registrant_privacy` | bool | No | WHOIS privacy for registrant (default: true) |
| `tech_privacy` | bool | No | WHOIS privacy for tech (default: true) |
| `nameservers` | list(string) | No | Custom nameservers |
| `allow_delete` | bool | No | Allow domain deletion on destroy (default: false) |
| `registration_timeout` | number | No | Registration timeout in seconds (default: 900) |

### Contact Object

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `first_name` | string | Yes | First name |
| `last_name` | string | Yes | Last name |
| `email` | string | Yes | Email address |
| `phone_number` | string | Yes | Phone in E.164 format (+1.5551234567) |
| `address_line_1` | string | Yes | Street address line 1 |
| `address_line_2` | string | No | Street address line 2 |
| `city` | string | Yes | City |
| `state` | string | Yes | State/province |
| `zip_code` | string | Yes | Postal code |
| `country_code` | string | Yes | Two-letter country code (US, UK, etc.) |
| `contact_type` | string | No | PERSON, COMPANY, ASSOCIATION, PUBLIC_BODY, or RESELLER |

### Attributes (Read-Only)

| Name | Type | Description |
|------|------|-------------|
| `id` | string | The domain name |
| `status` | string | Current domain status |
| `creation_date` | string | Domain creation date (RFC3339) |
| `expiration_date` | string | Domain expiration date (RFC3339) |

## AWS Permissions

The provider requires the following IAM permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "route53domains:RegisterDomain",
        "route53domains:GetDomainDetail",
        "route53domains:GetOperationDetail",
        "route53domains:UpdateDomainNameservers",
        "route53domains:UpdateDomainContact",
        "route53domains:UpdateDomainContactPrivacy",
        "route53domains:EnableDomainAutoRenew",
        "route53domains:DisableDomainAutoRenew",
        "route53domains:DeleteDomain",
        "route53domains:ListDomains"
      ],
      "Resource": "*"
    }
  ]
}
```

## Development

### Running Tests

```bash
# Unit tests (no AWS credentials required)
go test -v ./...

# Acceptance tests (requires AWS credentials and will register real domains)
TF_ACC=1 go test -v ./... -timeout 30m
```

### Building

```bash
go build -o terraform-provider-awsdomains
```

## License

MPL-2.0 - see [LICENSE](LICENSE)
