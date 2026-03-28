FROM golang:1.25-alpine AS builder

ENV SHELL=/bin/sh

RUN apk add --no-cache curl git libgcc libstdc++ ripgrep

RUN curl -fsSL https://opencode.ai/install | sh

ENV PATH="/root/.opencode/bin:${PATH}"

RUN git clone https://github.com/anomalyco/opencode-sdk-go.git /app/opencode-sdk-go

WORKDIR /app/opencode-sdk-go

CMD [ "opencode", "--port", "3000", "--hostname", "0.0.0.0" ]
