FROM golang:1.25-alpine AS builder

# ARG GH_USERNAME
# ARG GH_REPOSITORY

ENV SHELL=/bin/sh

RUN apk add --no-cache curl git libgcc libstdc++ ripgrep

RUN curl -fsSL https://opencode.ai/install | sh

ENV PATH="/root/.opencode/bin:${PATH}"

RUN git config --global user.email "sithumsandeepap@gmail.com" && git config --global user.name "Sithum Bopitiya"

WORKDIR /app/repository

CMD [ "opencode", "--port", "3000", "--hostname", "0.0.0.0" ]
