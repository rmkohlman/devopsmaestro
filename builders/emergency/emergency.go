// Package emergency provides the lightweight fallback container image used by
// `dvm attach --emergency`.
//
// The emergency image is intentionally minimal: an Alpine base with shell,
// git, common editors (vim, nano), curl/wget, ca-certificates, and a non-root
// `dev` user (UID/GID 1000) matching the standard dvm container user.
//
// It exists so users can still get into a workspace and fix things when the
// normal app image is broken — for example, a Dockerfile change that breaks
// the build, a corrupted lockfile, or a typo that prevents `dvm build`.
//
// The Dockerfile is embedded at compile time so the binary is self-contained
// and the image can be built lazily on first use without any external assets.
package emergency

import (
	_ "embed"
)

// ImageName is the canonical tag used for the cached emergency image.
// Bumping the suffix forces a rebuild on next `dvm attach --emergency`.
const ImageName = "dvm-emergency:v1"

// ContainerNamePrefix is used to label/name emergency containers so they
// can be discovered for cleanup independently of normal workspace containers.
const ContainerNamePrefix = "dvm-emergency-"

// LabelKey is the container label that identifies an emergency session.
// Used to enumerate orphaned emergency containers for cleanup.
const LabelKey = "dvm.emergency"

//go:embed Dockerfile
var dockerfileContents string

// Dockerfile returns the embedded Dockerfile contents used to build the
// emergency fallback image. Returning a copy keeps callers from depending on
// the embed mechanics.
func Dockerfile() string {
	return dockerfileContents
}
