FROM alpine:3.14.1

RUN apk add --no-cache ca-certificates

ADD ./apiworker /apiworker

ENTRYPOINT ["/apiworker"]
