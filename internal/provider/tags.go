package provider

import (
	"context"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	route53domainTypes "github.com/aws/aws-sdk-go-v2/service/route53domains/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	tftypes "github.com/hashicorp/terraform-plugin-framework/types"
)

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

func stringMapToFrameworkMap(tags map[string]string) (tftypes.Map, diag.Diagnostics) {
	values := make(map[string]attr.Value, len(tags))
	for key, value := range tags {
		values[key] = tftypes.StringValue(value)
	}

	return tftypes.MapValue(tftypes.StringType, values)
}

func cloneTags(tags map[string]string) map[string]string {
	cloned := make(map[string]string, len(tags))
	for key, value := range tags {
		cloned[key] = value
	}
	return cloned
}

func mergeTags(defaultTags, resourceTags map[string]string) map[string]string {
	merged := cloneTags(defaultTags)
	for key, value := range resourceTags {
		merged[key] = value
	}
	return merged
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

func tagKeysToDelete(currentTags, desiredTags map[string]string) []string {
	var keys []string
	for key := range currentTags {
		if _, ok := desiredTags[key]; !ok {
			keys = append(keys, key)
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
