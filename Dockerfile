# Build the manager binary
FROM golang:1.15-alpine as builder
RUN apk add --update --no-cache ca-certificates git

WORKDIR /workspace

# Better dep caching (other option is to vendor deps)
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -ldflags="-w -s" -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:latest
WORKDIR /
COPY --from=builder /workspace/manager .
ENTRYPOINT ["/manager"]
