package builders

import (
	"fmt"
	"strings"
)

// CICD builder stages — pinned upstream releases with SHA256 checksum verification.
// All four use Alpine base + curl, mirroring the lazygit/starship pattern.
//
// Arch detection on Alpine: `apk --print-arch` yields "x86_64" / "aarch64";
// we map to the URL conventions of each upstream project.
//
// Security policy: NEVER use `apk add kubectl/helm/kustomize` (Alpine community
// packages lag upstream and are not signed by upstream projects). Always pin to
// a specific upstream release and verify SHA256 (see #404 security review H1).

// generateKubectlBuilder downloads kubectl from kubernetes.io official release.
// URL: https://dl.k8s.io/release/vX.Y.Z/bin/linux/<arch>/kubectl
func (g *DefaultDockerfileGenerator) generateKubectlBuilder(df *strings.Builder) {
	df.WriteString("# --- Parallel builder: kubectl (pinned upstream + SHA256) ---\n")
	df.WriteString(fmt.Sprintf("FROM %s AS kubectl-builder\n", pinnedImage("alpine:3.20")))
	df.WriteString(g.apkCacheMountsLocked())
	df.WriteString("    set -e && \\\n")
	df.WriteString("    apk add --no-cache curl ca-certificates && \\\n")
	df.WriteString("    ARCH=$(apk --print-arch) && \\\n")
	df.WriteString("    if [ \"$ARCH\" = \"aarch64\" ]; then \\\n")
	df.WriteString(fmt.Sprintf("        K_ARCH=\"arm64\"; K_SHA256=\"%s\"; \\\n", kubectlChecksumArm64))
	df.WriteString("    elif [ \"$ARCH\" = \"x86_64\" ]; then \\\n")
	df.WriteString(fmt.Sprintf("        K_ARCH=\"amd64\"; K_SHA256=\"%s\"; \\\n", kubectlChecksumAmd64))
	df.WriteString("    else \\\n")
	df.WriteString("        echo \"ERROR: Unsupported architecture: $ARCH\"; exit 1; \\\n")
	df.WriteString("    fi && \\\n")
	df.WriteString(fmt.Sprintf("    curl %s -o /tmp/kubectl \"https://dl.k8s.io/release/v%s/bin/linux/${K_ARCH}/kubectl\" && \\\n", curlFlags, kubectlVersion))
	df.WriteString("    echo \"${K_SHA256}  /tmp/kubectl\" | sha256sum -c - && \\\n")
	df.WriteString("    install -m 0755 /tmp/kubectl /usr/local/bin/kubectl && \\\n")
	df.WriteString("    rm /tmp/kubectl && \\\n")
	df.WriteString("    test -x /usr/local/bin/kubectl\n\n")
}

// generateHelmBuilder downloads helm from get.helm.sh official release tarball.
// URL: https://get.helm.sh/helm-vX.Y.Z-linux-<arch>.tar.gz
func (g *DefaultDockerfileGenerator) generateHelmBuilder(df *strings.Builder) {
	df.WriteString("# --- Parallel builder: helm (pinned upstream + SHA256) ---\n")
	df.WriteString(fmt.Sprintf("FROM %s AS helm-builder\n", pinnedImage("alpine:3.20")))
	df.WriteString(g.apkCacheMountsLocked())
	df.WriteString("    set -e && \\\n")
	df.WriteString("    apk add --no-cache curl ca-certificates && \\\n")
	df.WriteString("    ARCH=$(apk --print-arch) && \\\n")
	df.WriteString("    if [ \"$ARCH\" = \"aarch64\" ]; then \\\n")
	df.WriteString(fmt.Sprintf("        H_ARCH=\"arm64\"; H_SHA256=\"%s\"; \\\n", helmChecksumArm64))
	df.WriteString("    elif [ \"$ARCH\" = \"x86_64\" ]; then \\\n")
	df.WriteString(fmt.Sprintf("        H_ARCH=\"amd64\"; H_SHA256=\"%s\"; \\\n", helmChecksumAmd64))
	df.WriteString("    else \\\n")
	df.WriteString("        echo \"ERROR: Unsupported architecture: $ARCH\"; exit 1; \\\n")
	df.WriteString("    fi && \\\n")
	df.WriteString(fmt.Sprintf("    curl %s -o /tmp/helm.tgz \"https://get.helm.sh/helm-v%s-linux-${H_ARCH}.tar.gz\" && \\\n", curlFlags, helmVersion))
	df.WriteString("    echo \"${H_SHA256}  /tmp/helm.tgz\" | sha256sum -c - && \\\n")
	df.WriteString("    tar -C /tmp -xzf /tmp/helm.tgz && \\\n")
	df.WriteString("    install -m 0755 /tmp/linux-${H_ARCH}/helm /usr/local/bin/helm && \\\n")
	df.WriteString("    rm -rf /tmp/helm.tgz /tmp/linux-${H_ARCH} && \\\n")
	df.WriteString("    test -x /usr/local/bin/helm\n\n")
}

// generateKustomizeBuilder downloads kustomize from GitHub release tarball.
// URL: https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize/vX.Y.Z/kustomize_vX.Y.Z_linux_<arch>.tar.gz
func (g *DefaultDockerfileGenerator) generateKustomizeBuilder(df *strings.Builder) {
	df.WriteString("# --- Parallel builder: kustomize (pinned upstream + SHA256) ---\n")
	df.WriteString(fmt.Sprintf("FROM %s AS kustomize-builder\n", pinnedImage("alpine:3.20")))
	df.WriteString(g.apkCacheMountsLocked())
	df.WriteString("    set -e && \\\n")
	df.WriteString("    apk add --no-cache curl ca-certificates && \\\n")
	df.WriteString("    ARCH=$(apk --print-arch) && \\\n")
	df.WriteString("    if [ \"$ARCH\" = \"aarch64\" ]; then \\\n")
	df.WriteString(fmt.Sprintf("        KU_ARCH=\"arm64\"; KU_SHA256=\"%s\"; \\\n", kustomizeChecksumArm64))
	df.WriteString("    elif [ \"$ARCH\" = \"x86_64\" ]; then \\\n")
	df.WriteString(fmt.Sprintf("        KU_ARCH=\"amd64\"; KU_SHA256=\"%s\"; \\\n", kustomizeChecksumAmd64))
	df.WriteString("    else \\\n")
	df.WriteString("        echo \"ERROR: Unsupported architecture: $ARCH\"; exit 1; \\\n")
	df.WriteString("    fi && \\\n")
	df.WriteString(fmt.Sprintf("    curl %s -o /tmp/kustomize.tgz \"https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize/v%s/kustomize_v%s_linux_${KU_ARCH}.tar.gz\" && \\\n", curlFlags, kustomizeVersion, kustomizeVersion))
	df.WriteString("    echo \"${KU_SHA256}  /tmp/kustomize.tgz\" | sha256sum -c - && \\\n")
	df.WriteString("    tar -C /tmp -xzf /tmp/kustomize.tgz kustomize && \\\n")
	df.WriteString("    install -m 0755 /tmp/kustomize /usr/local/bin/kustomize && \\\n")
	df.WriteString("    rm /tmp/kustomize.tgz /tmp/kustomize && \\\n")
	df.WriteString("    test -x /usr/local/bin/kustomize\n\n")
}

// generateArgoCDBuilder downloads the argocd CLI from GitHub release.
// URL: https://github.com/argoproj/argo-cd/releases/download/vX.Y.Z/argocd-linux-<arch>
// Conditional on .argocd/ detection (g.argoCDDetected).
func (g *DefaultDockerfileGenerator) generateArgoCDBuilder(df *strings.Builder) {
	df.WriteString("# --- Parallel builder: argocd CLI (pinned upstream + SHA256, conditional) ---\n")
	df.WriteString(fmt.Sprintf("FROM %s AS argocd-builder\n", pinnedImage("alpine:3.20")))
	df.WriteString(g.apkCacheMountsLocked())
	df.WriteString("    set -e && \\\n")
	df.WriteString("    apk add --no-cache curl ca-certificates && \\\n")
	df.WriteString("    ARCH=$(apk --print-arch) && \\\n")
	df.WriteString("    if [ \"$ARCH\" = \"aarch64\" ]; then \\\n")
	df.WriteString(fmt.Sprintf("        A_ARCH=\"arm64\"; A_SHA256=\"%s\"; \\\n", argocdChecksumArm64))
	df.WriteString("    elif [ \"$ARCH\" = \"x86_64\" ]; then \\\n")
	df.WriteString(fmt.Sprintf("        A_ARCH=\"amd64\"; A_SHA256=\"%s\"; \\\n", argocdChecksumAmd64))
	df.WriteString("    else \\\n")
	df.WriteString("        echo \"ERROR: Unsupported architecture: $ARCH\"; exit 1; \\\n")
	df.WriteString("    fi && \\\n")
	df.WriteString(fmt.Sprintf("    curl %s -o /tmp/argocd \"https://github.com/argoproj/argo-cd/releases/download/v%s/argocd-linux-${A_ARCH}\" && \\\n", curlFlags, argocdVersion))
	df.WriteString("    echo \"${A_SHA256}  /tmp/argocd\" | sha256sum -c - && \\\n")
	df.WriteString("    install -m 0755 /tmp/argocd /usr/local/bin/argocd && \\\n")
	df.WriteString("    rm /tmp/argocd && \\\n")
	df.WriteString("    test -x /usr/local/bin/argocd\n\n")
}

// generateStarshipBuilderAlpine emits an Alpine-based starship builder for the
// CICD path. The default generateStarshipBuilder uses debian:bookworm-slim
// (apt-get), which would contaminate the CICD image with Debian artifacts
// (#404). The starship release ships a musl-linux tarball that runs natively
// on Alpine, so we just swap the base image and use apk for curl.
func (g *DefaultDockerfileGenerator) generateStarshipBuilderAlpine(df *strings.Builder) {
	df.WriteString("# --- Parallel builder: Starship prompt (Alpine) ---\n")
	df.WriteString(fmt.Sprintf("FROM %s AS starship-builder\n", pinnedImage("alpine:3.20")))
	df.WriteString(g.apkCacheMountsLocked())
	df.WriteString("    set -e && \\\n")
	df.WriteString("    apk add --no-cache curl ca-certificates && \\\n")
	df.WriteString("    ARCH=$(apk --print-arch) && \\\n")
	df.WriteString("    if [ \"$ARCH\" = \"aarch64\" ]; then \\\n")
	df.WriteString(fmt.Sprintf("        STARSHIP_ARCH=\"aarch64-unknown-linux-musl\"; STARSHIP_SHA256=\"%s\"; \\\n", starshipChecksumArm64))
	df.WriteString("    elif [ \"$ARCH\" = \"x86_64\" ]; then \\\n")
	df.WriteString(fmt.Sprintf("        STARSHIP_ARCH=\"x86_64-unknown-linux-musl\"; STARSHIP_SHA256=\"%s\"; \\\n", starshipChecksumX86_64))
	df.WriteString("    else \\\n")
	df.WriteString("        echo \"ERROR: Unsupported architecture: $ARCH\"; exit 1; \\\n")
	df.WriteString("    fi && \\\n")
	df.WriteString(fmt.Sprintf("    curl %s -o /tmp/starship.tar.gz \"https://github.com/starship/starship/releases/download/v%s/starship-${STARSHIP_ARCH}.tar.gz\" && \\\n", curlFlags, starshipVersion))
	df.WriteString("    echo \"${STARSHIP_SHA256}  /tmp/starship.tar.gz\" | sha256sum -c - && \\\n")
	df.WriteString("    tar -C /usr/local/bin -xzf /tmp/starship.tar.gz starship && \\\n")
	df.WriteString("    rm /tmp/starship.tar.gz && \\\n")
	df.WriteString("    test -x /usr/local/bin/starship\n\n")
}
