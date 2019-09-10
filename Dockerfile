FROM golang:1.13.0-alpine AS build-env
LABEL intermediate=true

COPY ./ /home/build
WORKDIR /home/build
RUN go build -o /go/bin/runtime main.go

# Final stage
FROM alpine

RUN apk update
RUN apk add ca-certificates && rm -rf /var/cache/apk/*

ENV PORT=8080
ENV SCW_UPSTREAM_HOST="http://127.0.0.1"
ENV SCW_UPSTREAM_PORT=8081

WORKDIR /home/app
# Import built binary for core runtime
COPY --from=build-env /go/bin /home/app
ENTRYPOINT ["/home/app/runtime"]
