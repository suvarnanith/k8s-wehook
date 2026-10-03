# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:1.27.1-alpine AS builder
ARG TARGETOS=linux
ARG TARGETARCH=amd64
WORKDIR /src

COPY go.mod go.sum ./
COPY vendor ./vendor
COPY webhook ./webhook

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -mod=vendor -ldflags="-s -w" -o /out/webhook ./webhook

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /out/webhook /bin/webhook

# TLS cert/key are expected to be mounted at runtime, e.g. via a Secret volume
EXPOSE 9443

ENTRYPOINT ["/bin/webhook"]