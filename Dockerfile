FROM golang:1.25-alpine AS builder

ENV SHELL=/bin/sh

RUN apk add --no-cache curl git libgcc libstdc++ ripgrep

RUN curl -fsSL https://opencode.ai/install | sh

ENV PATH="/root/.opencode/bin:${PATH}"

RUN git config --global user.email "sithumsandeepap@gmail.com" && git config --global user.name "Sithum Bopitiya"

RUN cat <<'EOF' > /entrypoint.sh
#!/bin/sh
set -e

if [ -n "$GH_TOKEN" ]; then
    echo "==> GH_TOKEN exists, configuring git..."
    git config --global url."https://${GH_TOKEN}@github.com/".insteadOf "https://github.com/"
fi

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

exec "$@"
EOF

RUN chmod +x /entrypoint.sh

WORKDIR /app/repository

ENTRYPOINT ["/entrypoint.sh"]

CMD [ "opencode", "--port", "3000", "--hostname", "0.0.0.0" ]
