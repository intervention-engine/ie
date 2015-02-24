# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang

# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/intervention-engine/ie

WORKDIR /go/src/github.com/intervention-engine/ie
RUN go get
RUN go build server.go

# Run the outyet command by default when the container starts.
ENTRYPOINT /go/src/github.com/intervention-engine/ie/server

# Document that the service listens on port 3001.
EXPOSE 3001
