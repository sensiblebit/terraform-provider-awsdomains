package provider

import (
	"reflect"
	"testing"
)

func TestMergeTagsResourceOverridesDefaultTags(t *testing.T) {
	defaultTags := map[string]string{
		"Environment": "shared",
		"Owner":       "platform",
	}
	resourceTags := map[string]string{
		"Environment": "prod",
		"Name":        "example.com",
	}

	got := mergeTags(defaultTags, resourceTags)
	want := map[string]string{
		"Environment": "prod",
		"Owner":       "platform",
		"Name":        "example.com",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("mergeTags() = %#v, want %#v", got, want)
	}
}

func TestTagDiff(t *testing.T) {
	currentTags := map[string]string{
		"Keep":   "same",
		"Remove": "old",
		"Update": "old",
	}
	desiredTags := map[string]string{
		"Keep":   "same",
		"Update": "new",
		"Create": "new",
	}

	if got, want := tagKeysToDelete(currentTags, desiredTags), []string{"Remove"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("tagKeysToDelete() = %#v, want %#v", got, want)
	}

	gotUpdates := tagsToUpdate(currentTags, desiredTags)
	wantUpdates := map[string]string{
		"Create": "new",
		"Update": "new",
	}
	if !reflect.DeepEqual(gotUpdates, wantUpdates) {
		t.Fatalf("tagsToUpdate() = %#v, want %#v", gotUpdates, wantUpdates)
	}
}

func TestResourceTagsFromRemoteKeepsPriorResourceOverrides(t *testing.T) {
	remoteTags := map[string]string{
		"Environment": "shared",
		"Owner":       "platform",
		"Name":        "example.com",
	}
	priorResourceTags := map[string]string{
		"Environment": "shared",
		"Name":        "example.com",
	}

	got := resourceTagsFromRemote(remoteTags, priorResourceTags)
	want := map[string]string{
		"Environment": "shared",
		"Name":        "example.com",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("resourceTagsFromRemote() = %#v, want %#v", got, want)
	}
}

func TestResourceTagsFromRemoteIgnoresUntrackedRemoteTags(t *testing.T) {
	remoteTags := map[string]string{
		"External": "console",
		"Name":     "example.com",
	}
	priorResourceTags := map[string]string{
		"Name": "example.com",
	}

	got := resourceTagsFromRemote(remoteTags, priorResourceTags)
	want := map[string]string{
		"Name": "example.com",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("resourceTagsFromRemote() = %#v, want %#v", got, want)
	}
}
