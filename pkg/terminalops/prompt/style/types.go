// Package style provides types for PromptStyle resources.
// This package has ZERO dependencies on dvm internals and can be used standalone.
package style

import (
	"fmt"
	"regexp"
	"sort"
	"time"
)

// Kind constant for PromptStyle resource.
const KindPromptStyle = "PromptStyle"

// PromptStyle represents a terminal prompt style configuration.
// This is the canonical type used throughout terminal-manager.
type PromptStyle struct {
	// Core identification
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// Segments define the visual structure
	Segments []Segment `json:"segments,omitempty" yaml:"segments,omitempty"`

	// Suffix is appended to the format string
	Suffix string `json:"suffix,omitempty" yaml:"suffix,omitempty"`

	// Metadata
	Category string   `json:"category,omitempty" yaml:"category,omitempty"`
	Tags     []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	Enabled  bool     `json:"enabled" yaml:"enabled"`

	// Timestamps (optional, used when stored)
	CreatedAt *time.Time `json:"created_at,omitempty" yaml:"-"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" yaml:"-"`
}

// Segment represents a section of the prompt with visual styling.
type Segment struct {
	Name            string   `json:"name" yaml:"name"`
	Position        int      `json:"position" yaml:"position"`
	StartTransition string   `json:"startTransition,omitempty" yaml:"startTransition,omitempty"`
	StartColor      string   `json:"startColor,omitempty" yaml:"startColor,omitempty"`
	EndColor        string   `json:"endColor,omitempty" yaml:"endColor,omitempty"`
	Modules         []string `json:"modules,omitempty" yaml:"modules,omitempty"`
}

// PromptStyleYAML represents the full YAML document format (kubectl-style).
type PromptStyleYAML struct {
	APIVersion string              `yaml:"apiVersion"`
	Kind       string              `yaml:"kind"`
	Metadata   PromptStyleMetadata `yaml:"metadata"`
	Spec       PromptStyleSpec     `yaml:"spec"`
}

// PromptStyleMetadata contains style metadata in the YAML format.
type PromptStyleMetadata struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Category    string            `yaml:"category,omitempty"`
	Tags        []string          `yaml:"tags,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

// PromptStyleSpec contains the style specification in the YAML format.
type PromptStyleSpec struct {
	Segments []Segment `yaml:"segments,omitempty"`
	Suffix   string    `yaml:"suffix,omitempty"`
	Enabled  *bool     `yaml:"enabled,omitempty"`
}

// NewPromptStyle creates a new PromptStyle with default values.
func NewPromptStyle(name string) *PromptStyle {
	return &PromptStyle{
		Name:     name,
		Segments: make([]Segment, 0),
		Enabled:  true,
	}
}

// NewPromptStyleYAML creates a new PromptStyleYAML with proper defaults.
func NewPromptStyleYAML(name string) *PromptStyleYAML {
	return &PromptStyleYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       KindPromptStyle,
		Metadata: PromptStyleMetadata{
			Name: name,
		},
		Spec: PromptStyleSpec{
			Segments: make([]Segment, 0),
		},
	}
}

// ToPromptStyle converts PromptStyleYAML to the canonical PromptStyle type.
func (psy *PromptStyleYAML) ToPromptStyle() *PromptStyle {
	// Default to enabled unless explicitly set to false
	enabled := true
	if psy.Spec.Enabled != nil {
		enabled = *psy.Spec.Enabled
	}

	return &PromptStyle{
		Name:        psy.Metadata.Name,
		Description: psy.Metadata.Description,
		Category:    psy.Metadata.Category,
		Tags:        psy.Metadata.Tags,
		Segments:    psy.Spec.Segments,
		Suffix:      psy.Spec.Suffix,
		Enabled:     enabled,
	}
}

// GetSegments returns segments sorted by position.
func (ps *PromptStyle) GetSegments() []Segment {
	// Make a copy to avoid mutating the original
	segments := make([]Segment, len(ps.Segments))
	copy(segments, ps.Segments)

	// Sort by position
	sort.Slice(segments, func(i, j int) bool {
		return segments[i].Position < segments[j].Position
	})

	return segments
}

// Resource interface methods for PromptStyleYAML

// GetKind returns the resource kind.
func (psy *PromptStyleYAML) GetKind() string {
	return KindPromptStyle
}

// GetName returns the resource name.
func (psy *PromptStyleYAML) GetName() string {
	return psy.Metadata.Name
}

// GetAPIVersion returns the API version.
func (psy *PromptStyleYAML) GetAPIVersion() string {
	return psy.APIVersion
}

// Validate checks if the PromptStyleYAML resource is valid.
func (psy *PromptStyleYAML) Validate() error {
	if psy.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}

	if len(psy.Spec.Segments) == 0 {
		return fmt.Errorf("at least one segment is required")
	}

	// Check for duplicate positions
	positionMap := make(map[int]string)
	for _, seg := range psy.Spec.Segments {
		if existingSegment, exists := positionMap[seg.Position]; exists {
			return fmt.Errorf("duplicate position %d found in segments %q and %q", seg.Position, existingSegment, seg.Name)
		}
		positionMap[seg.Position] = seg.Name
	}

	// Check that colors are palette keys (not hex colors)
	hexColorRegex := regexp.MustCompile(`^#[0-9A-Fa-f]{3,8}$`)
	for _, seg := range psy.Spec.Segments {
		if seg.StartColor != "" && hexColorRegex.MatchString(seg.StartColor) {
			return fmt.Errorf("segment %q startColor %q must be palette key, not hex color", seg.Name, seg.StartColor)
		}
		if seg.EndColor != "" && hexColorRegex.MatchString(seg.EndColor) {
			return fmt.Errorf("segment %q endColor %q must be palette key, not hex color", seg.Name, seg.EndColor)
		}
	}

	return nil
}
