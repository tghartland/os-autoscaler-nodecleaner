FROM golang:alpine as builder
RUN apk add --no-cache git 
WORKDIR /build
COPY go.mod /build
RUN go mod download

FROM builder as compiler
WORKDIR /build
COPY *.go /build/
COPY .git /build/
RUN GIT_COMMIT=$(git rev-parse --short HEAD) && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main -ldflags "-X main.gitCommit=$GIT_COMMIT" .

FROM alpine
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=compiler /build/main /app
CMD ["/bin/sh"]
ENTRYPOINT ["/app/main"]
