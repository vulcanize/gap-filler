FROM golang:1.18-alpine as builder

RUN apk --update --no-cache add make git g++ linux-headers
# DEBUG
RUN apk add busybox-extras

# Get and build gap-filler
ADD . /go/src/github.com/vulcanize/gap-filler
WORKDIR /go/src/github.com/vulcanize/gap-filler
RUN make linux

# app container
FROM alpine

# keep binaries immutable
COPY --from=builder /go/src/github.com/vulcanize/gap-filler/build/gap-filler-linux /usr/local/bin/gap-filler

EXPOSE 8080

ENTRYPOINT ["gap-filler"]
