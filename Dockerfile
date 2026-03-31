FROM golang:1.25-alpine AS builder

ENV SHELL=/bin/sh

RUN apk add --no-cache curl git libgcc libstdc++ ripgrep

RUN curl -fsSL https://opencode.ai/install | sh

ENV PATH="/root/.opencode/bin:${PATH}"

RUN git config --global url."https://${GH_TOKEN}@github.com/".insteadOf "https://github.com/"

RUN git config --global user.email "sithumsandeepap@gmail.com" && git config --global user.name "Sithum Bopitiya"

RUN git clone https://${GH_TOKEN}@github.com/sithumonline/movie-box.git /app/movie-box

WORKDIR /app/movie-box

RUN go mod tidy

CMD [ "opencode", "--port", "3000", "--hostname", "0.0.0.0" ]
