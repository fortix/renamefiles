ARG DOCKER_HUB=registry-1.docker.io/
FROM ${DOCKER_HUB}library/golang:1.23.1-alpine AS builder

COPY . /src

RUN --mount=type=ssh \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
  apk update \
  && apk add make \
  && go env -w GOCACHE=/root/.cache/go-build \
  && go env -w GOMODCACHE=/go/pkg/mod \
  && go env -w GOPRIVATE=github.com/paularlott/ \
  \
  && cd /src \
  && go mod download \
  && go mod verify \
  && go mod tidy \
  \
  && make build

FROM ${DOCKER_HUB}library/alpine:3.20

# Upgrade to the latest versions
RUN apk update \
  && apk upgrade \
  && apk add bash

# Copy files in
COPY --from=builder /src/bin/proxydns /usr/local/bin/
COPY docker/proxydns.yml /etc/proxydns.yml

# Add a user, knot, to run the process
RUN addgroup -S proxydns \
  && adduser -S proxydns -G proxydns

# Set user and working directory
USER proxydns
WORKDIR /home/proxydns

# Set the entrypoint
CMD ["/usr/local/bin/proxydns"]
