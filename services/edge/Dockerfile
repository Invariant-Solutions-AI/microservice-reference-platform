# Copyright 2026 Invariant Solutions LLC
# SPDX-License-Identifier: Apache-2.0

ARG GO_VERSION=1.27.2

FROM golang:${GO_VERSION}-alpine AS build

WORKDIR /src

COPY go.work go.work.sum ./
COPY shared ./shared
COPY services/edge ./services/edge
COPY services/orders ./services/orders

WORKDIR /src/services/edge
RUN go mod download && \
    CGO_ENABLED=0 go build \
      -trimpath \
      -ldflags="-s -w" \
      -o /out/edge \
      ./cmd

FROM gcr.io/distroless/static-debian12:nonroot

ARG BUILD_DATE
ARG VCS_REF
LABEL org.opencontainers.image.title="edge" \
      org.opencontainers.image.description="Microservice reference platform edge service" \
      org.opencontainers.image.source="https://github.com/Invariant-Solutions-AI/microservice-reference-platform" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.revision="${VCS_REF}" \
      org.opencontainers.image.licenses="Apache-2.0"

COPY --from=build --chown=nonroot:nonroot /out/edge /edge

USER nonroot:nonroot
ENV EDGE_ADDR=:8080
EXPOSE 8080

ENTRYPOINT ["/edge"]
