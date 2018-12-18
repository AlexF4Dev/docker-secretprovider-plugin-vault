FROM golang:1.11-alpine AS build
ARG repo=gitlab.com/sirlatrom/docker-secretprovider-plugin-vault
WORKDIR /go/src
COPY vendor/ ./
RUN CGO_ENABLED=0 go install -v ./...
WORKDIR /go/src/$repo
COPY main.go .
RUN CGO_ENABLED=0 go install -v

FROM scratch
COPY --from=build "/go/bin/docker-secretprovider-plugin-vault" "/go/bin/docker-secretprovider-plugin-vault"
ENTRYPOINT ["/go/bin/docker-secretprovider-plugin-vault"]
