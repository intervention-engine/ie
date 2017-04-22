# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang:1.7

# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/intervention-engine/ie

#Build the Intervention Engine server
WORKDIR /go/src/github.com/intervention-engine/ie/cmd/ie
RUN go build main.go

# Document that the service listens on port 3001.
EXPOSE 3001

# Install Dockerize to get support for waiting on another container's port to be available.
# This is needed here so docker-compose can be configured to wait on the mongodb port to be available.
RUN apt-get update && apt-get install -y wget

ENV DOCKERIZE_VERSION v0.4.0
RUN wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && tar -C /usr/local/bin -xzvf dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && rm dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz
