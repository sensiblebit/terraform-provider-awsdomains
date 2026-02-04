---
page_title: "Provider: AWS Domains"
description: |-
  The AWS Domains provider is used to register and manage domain names through AWS Route53 Domains.
---

# AWS Domains Provider

The AWS Domains provider enables Terraform to register and manage domain names through [AWS Route53 Domains](https://aws.amazon.com/route53/features/#Domain_Registration).

## Example Usage

```terraform
terraform {
  required_providers {
    awsdomains = {
      source  = "sensiblebit/awsdomains"
      version = "~> 1.1"
    }
  }
}

provider "awsdomains" {
  region  = "us-east-1"
  profile = "default"
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

# Use the auto-created hosted zone directly - no data source needed!
resource "aws_route53_record" "www" {
  zone_id = awsdomains_domain.example.hosted_zone_id
  name    = "www.example.com"
  type    = "A"
  ttl     = 300
  records = ["192.0.2.1"]
}
```

## Authentication

The provider uses the AWS SDK for Go v2 and supports the standard AWS authentication methods:

- Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
- Shared credentials file (`~/.aws/credentials`)
- IAM roles for Amazon EC2

## Schema

### Optional

- `region` (String) AWS region. Must be `us-east-1` as Route53 Domains only operates in this region. Defaults to `us-east-1`.
- `profile` (String) AWS profile name from shared credentials file.

## Required IAM Permissions

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
        "route53domains:ListDomains",
        "route53domains:CheckDomainAvailability",
        "route53domains:ListPrices",
        "route53:ListHostedZonesByName",
        "route53:DeleteHostedZone"
      ],
      "Resource": "*"
    }
  ]
}
```
