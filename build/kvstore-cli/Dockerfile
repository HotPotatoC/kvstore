FROM golang:alpine AS build

ARG VERSION=unspecified
ARG GIT_COMMIT=unspecified
LABEL git_commit=$GIT_COMMIT

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

RUN mkdir -p /go/src/github.com/HotPotatoC/kvstore

ADD go.mod /go/src/github.com/HotPotatoC/kvstore
ADD go.sum /go/src/github.com/HotPotatoC/kvstore

WORKDIR /go/src/github.com/HotPotatoC/kvstore

RUN go mod download

COPY . .

RUN mkdir .bin

RUN go build -ldflags="-s -w -X 'main.Version=${VERSION}' -X 'main.Build=${GIT_COMMIT}'" -v -o \
    /go/bin/kvstore-cli cmd/kvstore-cli/main.go

FROM scratch

WORKDIR /usr/bin/

COPY --from=build /go/bin/kvstore-cli .

ENV PATH=/usr/bin/:$PATH

ENTRYPOINT [ "kvstore-cli" ]