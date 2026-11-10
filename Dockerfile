# OpenBooks on Chainguard images: a Wolfi-based toolchain for both build stages
# and cgr.dev/chainguard/static for the runtime, which is a like-for-like
# replacement for gcr.io/distroless/static (nothing but the CA bundle, tzdata
# and a nonroot uid). static carries its CA bundle at
# /etc/pki/tls/certs/ca-bundle.crt, which is one of the paths crypto/x509
# searches by default, so the TLS connection to IRC needs no extra files.
#
# These are the public cgr.dev/chainguard images, which publish :latest only.
# That is deliberate: the toolchain and the runtime should float forward onto
# rebuilds, and a toolchain version this repo actually cares about is already
# pinned by the go directive in go.mod. Nothing here comes from the private
# cgr.dev/jon.tutcher org, so the build needs no registry credentials.
#
# amd64 is what the cluster runs; nothing is fetched as a prebuilt binary, so
# the build works on any architecture Chainguard publishes.

# The client bundle. Built first and on its own so that the install layer is
# reused whenever only application source has changed.
FROM cgr.dev/chainguard/node:latest AS web
# The node image runs as uid 65532, but a WORKDIR is created root-owned, so
# npm could not write into it. Build stages are discarded, so root is fine.
USER root
WORKDIR /web
COPY server/app/package.json server/app/package-lock.json ./
RUN npm ci
COPY server/app/ ./
RUN npm run build

FROM cgr.dev/chainguard/go:latest AS build
WORKDIR /src
# Same idea: the module cache only has to be rebuilt when go.mod/go.sum change.
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Only the built client is needed; node_modules must not come across. The
# server embeds this directory with //go:embed, so it has to exist before the
# compile rather than being added to the image afterwards.
COPY --from=web /web/dist ./server/app/dist
ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags "-s -w" -o /openbooks ./cmd/openbooks

FROM cgr.dev/chainguard/static:latest AS app
ARG VERSION=dev
COPY --from=build /openbooks /usr/bin/openbooks

EXPOSE 80
VOLUME [ "/books" ]
ENV BASE_PATH=/

LABEL org.opencontainers.image.title="openbooks" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.source="https://gitlab.com/jontutcher/openbooks" \
      org.opencontainers.image.description="OpenBooks ${VERSION} on a Chainguard static base"

# --name is required by the CLI and deliberately not set here: every deployment
# supplies its own IRC nick as an argument.
ENTRYPOINT ["/usr/bin/openbooks", "server", "--dir", "/books", "--port", "80"]
