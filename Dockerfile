FROM golang:1.25-alpine

ENV SHELL=/bin/sh

RUN apk add --no-cache curl git libgcc libstdc++ ripgrep
RUN curl -fsSL https://opencode.ai/install | sh
ENV PATH="/root/.opencode/bin:${PATH}"

RUN git config --global user.email "sithumsandeepap@gmail.com" \
 && git config --global user.name "Sithum Bopitiya"

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

WORKDIR /app/repository
ENTRYPOINT ["/entrypoint.sh"]
CMD ["opencode", "--port", "3000", "--hostname", "0.0.0.0"]
