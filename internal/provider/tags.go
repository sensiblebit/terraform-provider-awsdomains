package provider

import (
	"context"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/aws/aws-sdk-go-v2/aws"
	route53domainTypes "github.com/aws/aws-sdk-go-v2/service/route53domains/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	tftypes "github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	maxDomainTags      = 50
	maxDomainTagKeyLen = 128
	maxDomainTagValLen = 256
)

var domainTagPattern = regexp.MustCompile(`^[A-Za-z0-9 .:/=+\-@]*$`)

func emptyFrameworkStringMap() tftypes.Map {
	return tftypes.MapValueMust(tftypes.StringType, map[string]attr.Value{})
}

func frameworkMapToStringMap(ctx context.Context, value tftypes.Map) (map[string]string, diag.Diagnostics) {
	tags := map[string]string{}
	if value.IsNull() || value.IsUnknown() {
		return tags, nil
	}

	diags := value.ElementsAs(ctx, &tags, false)
	return tags, diags
}

func frameworkMapElementsKnown(value tftypes.Map) bool {
	if value.IsUnknown() {
		return false
	}
	if value.IsNull() {
		return true
	}

	for _, element := range value.Elements() {
		if element.IsUnknown() {
			return false
		}
	}

	return true
}

func frameworkMapHasElements(value tftypes.Map) bool {
	if value.IsNull() || value.IsUnknown() {
		return false
	}
	return len(value.Elements()) > 0
}

func stringMapToFrameworkMap(tags map[string]string) (tftypes.Map, diag.Diagnostics) {
	values := make(map[string]attr.Value, len(tags))
	for key, value := range tags {
		values[key] = tftypes.StringValue(value)
	}

	return tftypes.MapValue(tftypes.StringType, values)
}

func cloneTags(tags map[string]string) map[string]string {
	cloned := make(map[string]string, len(tags))
	maps.Copy(cloned, tags)
	return cloned
}

func mergeTags(defaultTags, resourceTags map[string]string) map[string]string {
	merged := cloneTags(defaultTags)
	maps.Copy(merged, resourceTags)
	return merged
}

func tagManagementEnabled(defaultTags map[string]string, tagMaps ...tftypes.Map) bool {
	if len(defaultTags) > 0 {
		return true
	}

	return slices.ContainsFunc(tagMaps, frameworkMapHasElements)
}

func validateDomainTags(tags map[string]string) []string {
	var problems []string
	if len(tags) > maxDomainTags {
		problems = append(problems, fmt.Sprintf("tag count %d exceeds the Route53 Domains limit of %d", len(tags), maxDomainTags))
	}

	for key, value := range tags {
		keyLen := utf8.RuneCountInString(key)
		valueLen := utf8.RuneCountInString(value)

		switch {
		case keyLen == 0:
			problems = append(problems, "tag keys must be at least 1 character long")
		case keyLen > maxDomainTagKeyLen:
			problems = append(problems, fmt.Sprintf("tag key %q is %d characters; maximum is %d", key, keyLen, maxDomainTagKeyLen))
		}

		if valueLen > maxDomainTagValLen {
			problems = append(problems, fmt.Sprintf("tag value for key %q is %d characters; maximum is %d", key, valueLen, maxDomainTagValLen))
		}

		if !domainTagPattern.MatchString(key) {
			problems = append(problems, fmt.Sprintf("tag key %q contains invalid characters; allowed characters are letters, numbers, spaces, and .:/=+-@", key))
		}
		if !domainTagPattern.MatchString(value) {
			problems = append(problems, fmt.Sprintf("tag value for key %q contains invalid characters; allowed characters are letters, numbers, spaces, and .:/=+-@", key))
		}
	}

	sort.Strings(problems)
	return problems
}

func addTagValidationDiagnostics(diags *diag.Diagnostics, attributePath path.Path, tags map[string]string) {
	problems := validateDomainTags(tags)
	if len(problems) == 0 {
		return
	}

	diags.AddAttributeError(
		attributePath,
		"Invalid Route53 Domains Tags",
		strings.Join(problems, "\n"),
	)
}

func awsTagsToStringMap(tags []route53domainTypes.Tag) map[string]string {
	result := make(map[string]string, len(tags))
	for _, tag := range tags {
		result[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}
	return result
}

func stringMapToAWSTags(tags map[string]string) []route53domainTypes.Tag {
	keys := sortedTagKeys(tags)
	result := make([]route53domainTypes.Tag, 0, len(keys))

	for _, key := range keys {
		result = append(result, route53domainTypes.Tag{
			Key:   aws.String(key),
			Value: aws.String(tags[key]),
		})
	}

	return result
}

func sortedTagKeys(tags map[string]string) []string {
	keys := make([]string, 0, len(tags))
	for key := range tags {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func tagKeysToDelete(currentTags, desiredTags, previousManagedTags map[string]string) []string {
	var keys []string
	for key := range currentTags {
		if _, ok := desiredTags[key]; !ok {
			if _, managed := previousManagedTags[key]; managed {
				keys = append(keys, key)
			}
		}
	}
	sort.Strings(keys)
	return keys
}

func tagsToUpdate(currentTags, desiredTags map[string]string) map[string]string {
	result := map[string]string{}
	for key, desiredValue := range desiredTags {
		if currentValue, ok := currentTags[key]; !ok || currentValue != desiredValue {
			result[key] = desiredValue
		}
	}
	return result
}

func resourceTagsFromRemote(remoteTags, priorResourceTags map[string]string) map[string]string {
	resourceTags := map[string]string{}

	for key := range priorResourceTags {
		if value, ok := remoteTags[key]; ok {
			resourceTags[key] = value
		}
	}

	return resourceTags
}

func managedTagsFromRemote(remoteTags, managedTags map[string]string) map[string]string {
	tags := map[string]string{}

	for key := range managedTags {
		if value, ok := remoteTags[key]; ok {
			tags[key] = value
		}
	}

	return tags
}

func trackedTagKeys(defaultTags, resourceTags, previousManagedTags map[string]string) map[string]string {
	keys := make(map[string]string, len(defaultTags)+len(resourceTags)+len(previousManagedTags))
	for key := range defaultTags {
		keys[key] = ""
	}
	for key := range resourceTags {
		keys[key] = ""
	}
	for key := range previousManagedTags {
		keys[key] = ""
	}
	return keys
}
