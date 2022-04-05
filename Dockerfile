FROM alpine:3.15.4

RUN apk add --no-cache ca-certificates

ADD ./apiworker /apiworker

ENTRYPOINT ["/apiworker"]
