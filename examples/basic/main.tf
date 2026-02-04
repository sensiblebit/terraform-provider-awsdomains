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

variable "domain_name" {
  description = "Domain name to register"
  type        = string
}

variable "contact_email" {
  description = "Contact email for domain registration"
  type        = string
}

resource "awsdomains_domain" "example" {
  domain_name    = var.domain_name
  duration_years = 1
  auto_renew     = false

  admin_contact = {
    first_name     = "John"
    last_name      = "Doe"
    email          = var.contact_email
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
    email          = var.contact_email
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
    email          = var.contact_email
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

  # Safety: domain won't be deleted on terraform destroy
  allow_delete = false
}

output "domain_status" {
  description = "Current status of the domain"
  value       = awsdomains_domain.example.status
}

output "expiration_date" {
  description = "Domain expiration date"
  value       = awsdomains_domain.example.expiration_date
}

output "hosted_zone_id" {
  description = "Route53 hosted zone ID for the domain"
  value       = awsdomains_domain.example.hosted_zone_id
}
