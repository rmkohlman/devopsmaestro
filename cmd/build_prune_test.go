package cmd

import (
	"testing"
)

func TestImageRepo(t *testing.T) {
	tests := []struct {
		name      string
		imageName string
		want      string
	}{
		{
			name:      "standard image with timestamp tag",
			imageName: "dvm-dev-myapp:20260415-234218",
			want:      "dvm-dev-myapp",
		},
		{
			name:      "image with no tag returns full name",
			imageName: "dvm-dev-myapp",
			want:      "dvm-dev-myapp",
		},
		{
			name:      "image with multiple colons uses last index",
			imageName: "registry.example.com:5000/dvm-dev-myapp:20260415-234218",
			want:      "registry.example.com:5000/dvm-dev-myapp",
		},
		{
			name:      "empty string returns empty",
			imageName: "",
			want:      "",
		},
		{
			name:      "colon at start returns full name (idx=0 guard)",
			imageName: ":notvalid",
			want:      ":notvalid",
		},
		{
			name:      "workspace-style image name",
			imageName: "dvm-myworkspace-myapp:20260101-120000",
			want:      "dvm-myworkspace-myapp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := imageRepo(tt.imageName)
			if got != tt.want {
				t.Errorf("imageRepo(%q) = %q, want %q", tt.imageName, got, tt.want)
			}
		})
	}
}

// TestPruneOldImages_SkipsWhenNilPlatform verifies that pruneOldImages is a
// no-op when the platform is nil (e.g. early exit before platform detection).
func TestPruneOldImages_SkipsWhenNilPlatform(t *testing.T) {
	bc := &buildContext{
		platform:  nil,
		imageName: "dvm-dev-myapp:20260415-120000",
	}
	// Should not panic — just return early.
	bc.pruneOldImages()
}

// TestPruneOldImages_SkipsWhenEmptyImageName verifies that pruneOldImages is a
// no-op when imageName is empty.
func TestPruneOldImages_SkipsWhenEmptyImageName(t *testing.T) {
	bc := &buildContext{
		imageName: "",
	}
	// Should not panic — just return early.
	bc.pruneOldImages()
}

// TestImageRepo_RoundTrip verifies that applying imageRepo twice (to a result
// that has no tag) is idempotent.
func TestImageRepo_RoundTrip(t *testing.T) {
	original := "dvm-dev-myapp:20260415-234218"
	repo := imageRepo(original)
	// Applying again to a tag-less name should return the same value.
	if imageRepo(repo) != repo {
		t.Errorf("imageRepo is not idempotent: imageRepo(%q) = %q", repo, imageRepo(repo))
	}
}

// TestImageRepo_PreservesRepoForSameRepoCheck ensures that the repo extracted
// from a tagged image matches the repo extracted from another image with the
// same repo but a different tag — this is the core invariant used by
// pruneImagesForRepo when deciding which images to keep/prune.
func TestImageRepo_PreservesRepoForSameRepoCheck(t *testing.T) {
	keep := "dvm-dev-myapp:20260415-234218"
	old := "dvm-dev-myapp:20260101-000000"

	if imageRepo(keep) != imageRepo(old) {
		t.Errorf("expected same repo for keep=%q and old=%q, got %q vs %q",
			keep, old, imageRepo(keep), imageRepo(old))
	}
}

// TestImageRepo_DifferentReposNotEqual verifies that two images with different
// repo names do not share a repo prefix.
func TestImageRepo_DifferentReposNotEqual(t *testing.T) {
	a := "dvm-dev-myapp:20260415-234218"
	b := "dvm-dev-otherapp:20260415-234218"

	if imageRepo(a) == imageRepo(b) {
		t.Errorf("expected different repos for %q and %q, both returned %q",
			a, b, imageRepo(a))
	}
}
