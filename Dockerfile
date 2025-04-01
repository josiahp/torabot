FROM golang:1.24.1-alpine AS builder

RUN apk update

COPY . .

ENV GOCACHE=/go/cache

RUN --mount=type=cache,target=/go/cache \
    go build -o /go/bin/torabot

FROM golang:1.24.1-alpine AS runner

WORKDIR /data

COPY --from=builder --link /go/bin/torabot /go/bin/torabot

ENTRYPOINT [ "/go/bin/torabot" ]