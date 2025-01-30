FROM golang:alpine AS builder

RUN apk update

COPY . .

ENV GOCACHE=/go/cache

RUN --mount=type=cache,target=/go/cache \
    go build -o /go/bin/torabot

FROM golang:alpine AS runner

WORKDIR /data

COPY --from=builder --link /go/bin/torabot /go/bin/torabot

ENTRYPOINT [ "/go/bin/torabot" ]