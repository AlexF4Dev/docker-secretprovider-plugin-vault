FROM golang:1.11-alpine AS build
ARG repo=lspgitlab01.alm.brand.dk/absukl/secrets-plugin
WORKDIR /go/src
COPY vendor/ ./
RUN CGO_ENABLED=0 go install -v ./...
WORKDIR /go/src/$repo
COPY *.go .
RUN CGO_ENABLED=0 go install -v

FROM scratch
COPY --from=build "/go/bin/secrets-plugin" "/go/bin/secrets-plugin"
ENTRYPOINT ["/go/bin/secrets-plugin"]
