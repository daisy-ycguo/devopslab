# Use the official Golang image to create a build artifact.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
FROM golang:1.12 as builder

# Copy local code to the container image.
WORKDIR /go/src/github.com/daisy-ycguo/devopslab
COPY hello.go hello_test.go .

# Run unit tests
RUN CGO_ENABLED=0 GOOS=linux go test

# Build the code inside the container.
RUN CGO_ENABLED=0 GOOS=linux go build -v -o hello

# Use a Docker multi-stage build to create a lean production image.
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM alpine

# Copy the binary to the production image from the builder stage.
COPY --from=builder /go/src/github.com/daisy-ycguo/devopslab/hello /hello

# Service must listen to $PORT environment variable.
# This default value facilitates local development.
ENV PORT 8080

# Run the web service on container startup.
CMD ["/hello"]
