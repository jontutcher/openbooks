FROM node:24-alpine AS web
WORKDIR /web
# The manifests are copied on their own so that the install layer is reused
# whenever only application source has changed.
COPY server/app/package.json server/app/package-lock.json ./
RUN npm ci
COPY server/app/ ./
RUN npm run build

FROM golang:1.26-alpine AS build
WORKDIR /go/src/
# Same idea: the module cache only has to be rebuilt when go.mod/go.sum change.
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Only the built client is needed; node_modules must not come across.
COPY --from=web /web/dist ./server/app/dist

ENV CGO_ENABLED=0
WORKDIR /go/src/cmd/openbooks/
RUN go build

FROM gcr.io/distroless/static AS app
WORKDIR /app
COPY --from=build /go/src/cmd/openbooks/openbooks .

EXPOSE 80
VOLUME [ "/books" ]
ENV BASE_PATH=/

ENTRYPOINT ["./openbooks", "server", "--dir", "/books", "--port", "80"]
