FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY . .
RUN go build -ldflags="-s -w" -o pghealthy .

FROM scratch
LABEL maintainer="floholz"
LABEL org.opencontainers.image.source="https://github.com/floholz/pghealthy"
LABEL org.opencontainers.image.description="PostgreSQL health check service"
LABEL org.opencontainers.image.title="pghealthy"
LABEL org.opencontainers.image.url="https://github.com/floholz/pghealthy"
LABEL org.opencontainers.image.vendor="floholz"
LABEL org.opencontainers.image.authors="floholz"
LABEL org.opencontainers.image.licenses="MIT"
COPY --from=builder /src/pghealthy /pghealthy
EXPOSE 2345
ENTRYPOINT ["/pghealthy"]
