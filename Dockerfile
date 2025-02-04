# Build the manager binary
FROM golang:1.13.1 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY pkg/ pkg/
COPY version/ version/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o manager main.go

FROM alpine:3.10
LABEL MAINTAINER 'Kazumichi Yamamoto <yamamoto.febc@gmail.com>'
# install xorriso to use mkisofs
RUN apk add --update --no-cache ca-certificates xorriso

WORKDIR /
COPY --from=builder /workspace/manager .

ENTRYPOINT ["/manager"]
