# Start with a full-fledged golang image, but strip it to the final image.
FROM golang:1.15-alpine as builder


COPY . /go/src/github.com/gdotgordon/fibsrv

WORKDIR /go/src/github.com/gdotgordon/fibsrv

RUN go build -v

FROM alpine:latest

WORKDIR /root/

# Make a significantly slimmed-down final result.
COPY --from=builder /go/src/github.com/gdotgordon/fibsrv .

LABEL maintainer="Gary Gordon <gagordon12@gmail.com>"

CMD ["./fibsrv"]