# This dockerfile for any release name defined in CI_TMPL_HELM_RELEASE_NAMES without it's own Dockerfile.<release_name>

# first build stage
# by default uses alpine-based image
FROM golang:alpine as builder
ARG VERSION

# copy all source code for build
WORKDIR /src
RUN go env -w GOCACHE=/cache/go-cache
RUN go env -w GOMODCACHE=/cache/gomod-cache
COPY ./go.* ./
RUN --mount=type=cache,target=/gomod-cache \
  go mod download
COPY . .

RUN  --mount=type=cache,target=/cache/gomod-cache --mount=type=cache,target=/cache/go-cache \
    CGO_ENABLED=0 GOOS=linux go build -mod=vendor -a -installsuffix cgo -o app -ldflags "-X 'main.version=${VERSION}'" ./cmd/migration

CMD ["./app", "show-config"]