FROM golang:1.25-alpine AS builder

# ARG GH_USERNAME
# ARG GH_REPOSITORY

ENV SHELL=/bin/sh

RUN apk add --no-cache curl git libgcc libstdc++ ripgrep

RUN curl -fsSL https://opencode.ai/install | sh

ENV PATH="/root/.opencode/bin:${PATH}"

RUN --mount=type=secret,id=GH_TOKEN,target=/run/secrets/GH_TOKEN \
    export GH_TOKEN=$(cat /run/secrets/GH_TOKEN) && \
    git config --global url."https://${GH_TOKEN}@github.com/".insteadOf "https://github.com/"

RUN git config --global user.email "sithumsandeepap@gmail.com" && git config --global user.name "Sithum Bopitiya"

RUN --mount=type=secret,id=GH_TOKEN,target=/run/secrets/GH_TOKEN \
    --mount=type=secret,id=GH_USERNAME,target=/run/secrets/GH_USERNAME \
    --mount=type=secret,id=GH_REPOSITORY,target=/run/secrets/GH_REPOSITORY \
    export GH_TOKEN=$(cat /run/secrets/GH_TOKEN) && \
    export GH_USERNAME=$(cat /run/secrets/GH_USERNAME) && \
    export GH_REPOSITORY=$(cat /run/secrets/GH_REPOSITORY) && \
    git clone https://${GH_TOKEN}@github.com/${GH_USERNAME}/${GH_REPOSITORY}.git /app/repository

WORKDIR /app/repository

RUN go mod tidy

CMD [ "opencode", "--port", "3000", "--hostname", "0.0.0.0" ]
