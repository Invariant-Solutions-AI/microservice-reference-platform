# Copyright 2026 Invariant Solutions LLC
# SPDX-License-Identifier: Apache-2.0

ARG GO_VERSION=1.27.3

FROM golang:${GO_VERSION}-alpine AS build

WORKDIR /src

COPY go.work go.work.sum ./
COPY shared ./shared
COPY services ./services

WORKDIR /src/services/orders
RUN go mod download && \
    CGO_ENABLED=0 go build \
      -trimpath \
      -ldflags="-s -w" \
      -o /out/orders \
      ./cmd

FROM gcr.io/distroless/static-debian12:nonroot

ARG BUILD_DATE
ARG VCS_REF
LABEL org.opencontainers.image.title="orders" \
      org.opencontainers.image.description="Microservice reference platform orders service" \
      org.opencontainers.image.source="https://github.com/Invariant-Solutions-AI/microservice-reference-platform" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.revision="${VCS_REF}" \
      org.opencontainers.image.licenses="Apache-2.0"

COPY --from=build --chown=nonroot:nonroot /out/orders /orders

USER nonroot:nonroot
ENV ORDERS_ADDR=:8081
EXPOSE 8081

ENTRYPOINT ["/orders"]
