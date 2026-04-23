# Lightweight emergency fallback container for `dvm attach --emergency`.
#
# Goals:
#   - Build in well under 30 seconds on a warm cache
#   - Provide just enough tooling for emergency edits (shell, git, editors)
#   - Include ca-certificates so corporate certs mounted at runtime work
#   - Run as a non-root `dev` user that matches the standard dvm UID/GID (1000:1000)
FROM alpine:3.20

# Essential tools only — keep the layer small and fast to build.
#   - bash:           more familiar interactive shell than ash
#   - git:            version control for emergency commits
#   - vim, nano:      editors (covers both camps)
#   - less:           pager for git, man, etc.
#   - curl, wget:     fetch dependencies / patches
#   - ca-certificates: trust store; corporate certs can be added at runtime
#   - openssh-client: needed for git over ssh
#   - tini:           proper PID 1 init so signals/zombies are handled
RUN apk add --no-cache \
        bash \
        git \
        vim \
        nano \
        less \
        curl \
        wget \
        ca-certificates \
        openssh-client \
        tini \
    && update-ca-certificates 2>/dev/null || true

# Standard dvm container user (UID/GID 1000:1000) so file ownership lines up
# with bind-mounted workspace files written by the host user.
RUN addgroup -g 1000 dev \
    && adduser -D -u 1000 -G dev -s /bin/bash dev \
    && mkdir -p /workspace \
    && chown dev:dev /workspace

# Make it visually obvious the user is in degraded/emergency mode.
RUN printf '%s\n' \
        'export PS1="\[\e[1;31m\][EMERGENCY]\[\e[0m\] \w \$ "' \
        'export DVM_EMERGENCY=1' \
        'echo' \
        'echo "  ⚠  You are in the dvm EMERGENCY fallback container."' \
        'echo "     Tools available: bash, git, vim, nano, curl, wget"' \
        'echo "     Workspace is mounted at /workspace"' \
        'echo "     Exit (Ctrl+D) when done; the container is ephemeral."' \
        'echo' \
        > /etc/profile.d/dvm-emergency.sh \
    && chmod +x /etc/profile.d/dvm-emergency.sh

USER dev
WORKDIR /workspace

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/bin/sleep", "infinity"]
