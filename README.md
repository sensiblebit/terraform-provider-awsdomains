# Terraform Provider for AWS Route53 Domain Registration

A Terraform provider for registering and managing domain names through AWS Route53 Domains.

## Quick Start

```hcl
terraform {
  required_providers {
    awsdomains = {
      source  = "sensiblebit/awsdomains"
      version = "~> 1.0"
    }
  }
}

provider "awsdomains" {
  region = "us-east-1"  # Required: Route53 Domains only works in us-east-1
}

resource "awsdomains_domain" "example" {
  domain_name    = "example.com"
  duration_years = 1

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
  }

  registrant_contact = { /* same structure */ }
  tech_contact       = { /* same structure */ }
}

# Use the auto-created hosted zone directly
resource "aws_route53_record" "apex" {
  zone_id = awsdomains_domain.example.hosted_zone_id
  name    = "example.com"
  type    = "A"
  # ...
}
```

## Features

- Register new domains with AWS Route53 Domains
- Manage domain contacts (admin, registrant, tech)
- Configure WHOIS privacy protection
- Update nameservers
- Manage auto-renewal settings
- **Auto-exposes `hosted_zone_id`** - no data source lookup needed
- Import existing domains into Terraform state
- Safe defaults: domains are NOT deleted on `terraform destroy` unless explicitly enabled

## Resource: awsdomains_domain

### Arguments

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `domain_name` | string | Yes | - | Domain name to register |
| `duration_years` | number | No | `1` | Years to register (1-10) |
| `auto_renew` | bool | No | `false` | Enable auto-renewal |
| `admin_contact` | object | Yes | - | Administrative contact |
| `registrant_contact` | object | Yes | - | Registrant contact |
| `tech_contact` | object | Yes | - | Technical contact |
| `admin_privacy` | bool | No | `true` | WHOIS privacy for admin |
| `registrant_privacy` | bool | No | `true` | WHOIS privacy for registrant |
| `tech_privacy` | bool | No | `true` | WHOIS privacy for tech |
| `nameservers` | list(string) | No | - | Custom nameservers |
| `allow_delete` | bool | No | `false` | Allow domain deletion on destroy |
| `delete_hosted_zone` | bool | No | `false` | Delete auto-created hosted zone (for external DNS) |
| `registration_timeout` | number | No | `900` | Timeout in seconds |

### Attributes (Read-Only)

| Name | Description |
|------|-------------|
| `id` | The domain name |
| `status` | Current domain status |
| `creation_date` | Domain creation date (RFC3339) |
| `expiration_date` | Domain expiration date (RFC3339) |
| `hosted_zone_id` | Route53 hosted zone ID (auto-created by AWS) |

### Contact Object

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `first_name` | string | Yes | First name |
| `last_name` | string | Yes | Last name |
| `email` | string | Yes | Email address |
| `phone_number` | string | Yes | E.164 format (+1.5551234567) |
| `address_line_1` | string | Yes | Street address |
| `address_line_2` | string | No | Street address line 2 |
| `city` | string | Yes | City |
| `state` | string | Yes | State/province |
| `zip_code` | string | Yes | Postal code |
| `country_code` | string | Yes | Two-letter code (US, UK, etc.) |
| `contact_type` | string | No | PERSON, COMPANY, ASSOCIATION, PUBLIC_BODY, RESELLER |

## Data Sources

### awsdomains_domain_availability

Check if a domain is available (free API).

```hcl
data "awsdomains_domain_availability" "check" {
  domain_name = "example.com"
}

output "available" {
  value = data.awsdomains_domain_availability.check.available
}
```

| Attribute | Type | Description |
|-----------|------|-------------|
| `domain_name` | string | Domain to check |
| `availability` | string | AVAILABLE, UNAVAILABLE, etc. |
| `available` | bool | True if registrable |

### awsdomains_domain_price

Get TLD pricing (free API).

```hcl
data "awsdomains_domain_price" "com" {
  tld = "com"
}

output "cost" {
  value = "${data.awsdomains_domain_price.com.registration_price} ${data.awsdomains_domain_price.com.currency}"
}
```

| Attribute | Type | Description |
|-----------|------|-------------|
| `tld` | string | Top-level domain |
| `registration_price` | number | Registration cost |
| `renewal_price` | number | Renewal cost |
| `transfer_price` | number | Transfer cost |
| `currency` | string | Currency code (USD) |

## Import

```bash
terraform import 'awsdomains_domain.example' example.com
```

**Note**: Contact information is NOT populated during import. First `apply` after import will set contacts.

---

# Technical Reference

## Architecture

### Provider Framework

Uses **Terraform Plugin Framework** (not SDK v2):
- `github.com/hashicorp/terraform-plugin-framework`
- Schema via structs with `tfsdk` tags
- Resources implement `resource.Resource` interface
- Data sources implement `datasource.DataSource` interface

### File Structure

```
internal/provider/
├── provider.go                      # Provider config, AWS client setup
├── domain_registration_resource.go  # Main resource (CRUD for domains)
├── domain_availability_data_source.go  # Free API
└── domain_price_data_source.go      # Free API
```

### AWS Clients

Provider creates two clients via `ProviderData` struct:
- `DomainsClient`: `*route53domains.Client` - domain registration operations
- `Route53Client`: `*route53.Client` - hosted zone lookups

**Region restriction**: Route53 Domains API only works in `us-east-1`

## Resource Lifecycle

### Create
1. `RegisterDomain` API call
2. Poll `GetOperationDetail` until `SUCCESSFUL` or timeout
3. `UpdateDomainNameservers` if specified
4. `GetDomainDetail` to fetch computed fields
5. If `delete_hosted_zone = true`: safely delete the registrar-created zone
6. Otherwise: `ListHostedZonesByName` to get hosted zone ID

### Read
1. `GetDomainDetail` API call
2. If error, removes resource from state (known issue - should distinguish 404)
3. `ListHostedZonesByName` to refresh hosted zone ID

### Update
1. `EnableDomainAutoRenew` / `DisableDomainAutoRenew` if changed
2. `UpdateDomainNameservers` if changed
3. `UpdateDomainContact` for contact changes
4. `UpdateDomainContactPrivacy` for privacy settings
5. Refresh state via `GetDomainDetail`

### Delete
- `allow_delete = false` (default): removes from state only, domain persists
- `allow_delete = true`: calls `DeleteDomain` API (may fail for some TLDs), then attempts to delete the hosted zone (best-effort, warns if zone has records)

### Import
Uses `ImportStatePassthroughID` setting both `domain_name` and `id`.

## AWS API Reference

### Free Operations
- `CheckDomainAvailability` - check availability
- `ListPrices` - TLD pricing
- `ListDomains` - list owned domains
- `GetDomainDetail` - domain details
- `ListHostedZonesByName` - find hosted zones

### Paid Operations
- `RegisterDomain` - ~$12-35+ per TLD
- `RenewDomain` - same as registration
- `TransferDomain` - varies

### May Fail
- `DeleteDomain` - not supported by all registries

## AWS Permissions

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
        "route53domains:CheckDomainAvailability",
        "route53domains:ListPrices",
        "        "route53:ListHostedZonesByName",
        "route53:ListResourceRecordSets",
        "route53:DeleteHostedZone""
      ],
      "Resource": "*"
    }
  ]
}
```

## Testing

### Unit Tests (no AWS required)
```bash
go test -v ./...
```

### Acceptance Tests

**Free API tests** (safe, no cost):
```bash
TF_ACC=1 go test -v ./... -run 'TestAccDomain(Availability|Price)'
```

**Full resource tests** (EXPENSIVE - registers real domains):
```bash
TF_ACC=1 go test -v ./... -run 'TestAccDomainRegistration' -timeout 30m
```

### Mock Client Pattern

Current implementation uses concrete `*route53domains.Client`. To enable mocking:
1. Define interface with required methods
2. Have resource accept interface
3. Inject mock in tests

## Common Issues

### "Cannot import non-existent remote object"
- Provider not configured
- AWS credentials missing/wrong profile
- Domain doesn't exist in account

Debug: `aws route53domains get-domain-detail --domain-name example.com --region us-east-1`

### "Invalid for_each argument"
`for_each` with dynamic values (like `plantimestamp()`) fails at import. Workaround:
1. Hardcode domain set
2. Import
3. Revert to dynamic

### Contact validation errors
Phone must be E.164: `+1.5551234567`

## Development

### Build
```bash
go build -o terraform-provider-awsdomains
```

### Local Testing
```hcl
# ~/.terraformrc
provider_installation {
  dev_overrides {
    "sensiblebit/awsdomains" = "/path/to/terraform-provider-awsdomains"
  }
  direct {}
}
```

### Dependencies
```bash
go get -u ./... && go mod tidy
```

## Publishing

1. Public repo named `terraform-provider-{NAME}`
2. GPG signing required
3. GitHub secrets:
   - `GPG_PRIVATE_KEY`
   - `PASSPHRASE`
4. Tag and push: `git tag v1.0.0 && git push origin v1.0.0`
5. GoReleaser builds multi-platform binaries

## Code Patterns

- `tftypes` alias for `github.com/hashicorp/terraform-plugin-framework/types`
- `contactModelToAWS()` converts TF models to AWS types
- `planmodifier.UseStateForUnknown()` for computed fields
- `stringplanmodifier.RequiresReplace()` for immutable fields
- `ProviderData` struct passes multiple clients to resources

## Future Improvements

1. **Better error handling in Read**: Distinguish 404 from other errors
2. **Interface for AWS client**: Enable proper unit testing with mocks
3. **Data source for listing owned domains**: `awsdomains_domains` (plural)
4. **Support for domain transfer**: `TransferDomain` API
5. **DNSSEC support**: `AssociateDelegationSignerToDomain` API

## License

MPL-2.0
