#!/bin/sh
set -e

# ── Git auth ──────────────────────────────────────────────
if [ -n "$GH_TOKEN" ]; then
    echo "==> GH_TOKEN exists, configuring git..."
    git config --global url."https://${GH_TOKEN}@github.com/".insteadOf "https://github.com/"
fi

# ── Repo setup ────────────────────────────────────────────
REPO_DIR="/app/repository"

if [ -d "${REPO_DIR}/.git" ]; then
    echo "==> Repository exists, updating..."
    DEFAULT_BRANCH=$(git -C "$REPO_DIR" remote show origin \
        | grep "HEAD branch" \
        | awk '{print $NF}')
    git -C "$REPO_DIR" checkout "$DEFAULT_BRANCH"
    git -C "$REPO_DIR" pull --ff-only
else
    echo "==> Cloning repository... $REPO_DIR"
    git clone "https://github.com/${GH_USERNAME}/${GH_REPOSITORY}.git" "$REPO_DIR"
fi

# Ensure OpenCode data dir exists and is writable
mkdir -p /root/.local/share/opencode
chmod 700 /root/.local/share/opencode || true

echo "==> Debug opencode data dir:"
ls -la /root/.local || true
ls -la /root/.local/share || true
ls -la /root/.local/share/opencode || true
id || true

exec "$@"
