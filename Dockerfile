ARG GO_VERSION
FROM golang:${GO_VERSION}-alpine AS builder

RUN apk update

COPY . .

ENV GOCACHE=/go/cache

RUN --mount=type=cache,target=/go/cache \
    go build -o /go/bin/torabot

FROM golang:${GO_VERSION}-alpine AS runner

WORKDIR /data

COPY --from=builder --link /go/bin/torabot /go/bin/torabot

ENTRYPOINT [ "/go/bin/torabot" ]