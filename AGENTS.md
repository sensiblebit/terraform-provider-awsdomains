# AGENTS.md - Developer & AI Agent Reference

Technical reference for contributors and AI agents working on this codebase.

## Architecture

### Provider Framework

This provider uses **Terraform Plugin Framework** (not the older SDK v2). Key differences:
- Uses `github.com/hashicorp/terraform-plugin-framework`
- Schema defined via structs with `tfsdk` tags
- Resources implement `resource.Resource` interface
- Data sources implement `datasource.DataSource` interface

### File Structure

```
internal/provider/
├── provider.go                      # Provider configuration, AWS client setup
├── domain_registration_resource.go  # Main resource (CRUD for domains)
├── domain_availability_data_source.go  # Read-only, free API
└── domain_price_data_source.go      # Read-only, free API
```

### AWS Client

- Client type: `*route53domains.Client` from AWS SDK v2
- Passed via `resp.ResourceData` / `resp.DataSourceData` in provider's `Configure()`
- Resources retrieve it in their own `Configure()` method
- **Region restriction**: Route53 Domains API only works in `us-east-1`

## Resource Lifecycle

### awsdomains_domain

**Create flow:**
1. `RegisterDomain` API call
2. Poll `GetOperationDetail` until status is `SUCCESSFUL` or timeout
3. Update nameservers if specified via `UpdateDomainNameservers`
4. Fetch final state via `GetDomainDetail`

**Read flow:**
1. `GetDomainDetail` API call
2. If error (any error, not just 404), removes resource from state
3. This is a known issue - should distinguish "not found" from other errors

**Update flow:**
1. `EnableDomainAutoRenew` / `DisableDomainAutoRenew` if changed
2. `UpdateDomainNameservers` if changed
3. `UpdateDomainContact` for contact changes
4. `UpdateDomainContactPrivacy` for privacy settings
5. Refresh state via `GetDomainDetail`

**Delete flow:**
1. If `allow_delete = false` (default): just remove from Terraform state, domain persists
2. If `allow_delete = true`: call `DeleteDomain` API (may fail for some TLDs)

### Import

Import uses `ImportStatePassthroughID` setting both `domain_name` and `id` to the import ID.

```bash
terraform import 'awsdomains_domain.example["foo.com"]' foo.com
```

**Known limitation**: Contact information is NOT populated during import. User must have contacts defined in their config, and first `apply` after import will set them.

## Testing Strategy

### Unit Tests (no AWS required)

Test schema, metadata, and helper functions:
```bash
go test -v ./...
```

### Acceptance Tests

**Free API tests** (CheckDomainAvailability, ListPrices):
```bash
TF_ACC=1 go test -v ./... -run 'TestAccDomain(Availability|Price)'
```

**Full resource tests** (EXPENSIVE - registers real domains):
```bash
TF_ACC=1 go test -v ./... -run 'TestAccDomainRegistration'
```

### Mock Client Pattern

For unit testing resource logic, create a mock implementing needed methods:

```go
type MockRoute53DomainsClient struct {
    GetDomainDetailFunc func(...) (*route53domains.GetDomainDetailOutput, error)
    // ... other methods
}
```

Currently, the resource directly uses `*route53domains.Client`. To enable mocking, would need to:
1. Define an interface with required methods
2. Have resource accept interface instead of concrete client
3. Inject mock in tests

## AWS API Reference

### Free Operations (use for testing)
- `CheckDomainAvailability` - check if domain can be registered
- `ListPrices` - get TLD pricing
- `ListDomains` - list registered domains (your account)
- `GetDomainSuggestions` - get alternative domain suggestions
- `GetDomainDetail` - details for domains you own

### Paid Operations
- `RegisterDomain` - ~$12-35+ depending on TLD
- `RenewDomain` - same as registration price
- `TransferDomain` - varies by TLD

### Operations That May Fail
- `DeleteDomain` - not supported by all registries, may return error

## Common Issues

### "Cannot import non-existent remote object"

Usually means:
1. Provider not configured (check `provider "awsdomains" {}` block)
2. AWS credentials missing or wrong profile
3. Domain doesn't exist in this AWS account

The Read function swallows all errors and removes from state. Check AWS CLI:
```bash
aws route53domains get-domain-detail --domain-name example.com --region us-east-1
```

### "Invalid for_each argument"

When using `for_each` with dynamic values (like `plantimestamp()`), Terraform can't determine keys at import time. Workaround:
1. Temporarily hardcode the domain set
2. Run import
3. Revert to dynamic values

### Contact validation errors

Phone must be E.164 format: `+1.5551234567` (country code, dot, number)

## Publishing to Terraform Registry

1. Repository must be public and named `terraform-provider-{NAME}`
2. Requires GPG signing for releases
3. Set up secrets in GitHub:
   - `GPG_PRIVATE_KEY` - ASCII armored private key
   - `PASSPHRASE` - GPG key passphrase
   - `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` - for acceptance tests

4. Tag and push:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

5. GoReleaser builds multi-platform binaries automatically

## Code Style

- Use `tftypes` alias for `github.com/hashicorp/terraform-plugin-framework/types`
- Helper functions like `contactModelToAWS()` convert between TF and AWS types
- Computed fields use `planmodifier.UseStateForUnknown()` to preserve across plans
- Required-replace fields use `stringplanmodifier.RequiresReplace()`

## Future Improvements

1. **Better error handling in Read**: Distinguish 404 from other errors
2. **Interface for AWS client**: Enable proper unit testing with mocks
3. **Data source for listing owned domains**: `awsdomains_domains` (plural)
4. **Support for domain transfer**: `TransferDomain` API
5. **DNSSEC support**: `AssociateDelegationSignerToDomain` API
